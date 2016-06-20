package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/drone/drone-plugin-go/plugin"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	buildCommit string
)

type terraform struct {
	Remote      remote            `json:"remote"`
	Plan        bool              `json:"plan"`
	Vars        map[string]string `json:"vars"`
	Cacert      string            `json:"ca_cert"`
	Sensitive   bool              `json:"sensitive"`
	RoleARN     string            `json:"role_arn_to_assume"`
	RootDir     string            `json:"root_dir"`
	Parallelism int               `json:"parallelism"`
}

type remote struct {
	Backend string            `json:"backend"`
	Config  map[string]string `json:"config"`
}

func main() {
	fmt.Printf("Drone Terraform Plugin built from %s\n", buildCommit)

	workspace := plugin.Workspace{}
	vargs := terraform{}

	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	if vargs.RoleARN != "" {
		assumeRole(vargs.RoleARN)
	}

	var commands []*exec.Cmd
	remote := vargs.Remote
	if vargs.Cacert != "" {
		commands = append(commands, installCaCert(vargs.Cacert))
	}
	if remote.Backend != "" {
		commands = append(commands, deleteCache())
		commands = append(commands, remoteConfigCommand(remote))
	}
	commands = append(commands, getModules())
	commands = append(commands, planCommand(vargs.Vars, vargs.Parallelism))
	if !vargs.Plan {
		commands = append(commands, applyCommand(vargs.Parallelism))
	}
	commands = append(commands, deleteCache())

	for _, c := range commands {
		c.Env = os.Environ()
		c.Dir = workspace.Path
		if c.Dir == "" {
			wd, err := os.Getwd()
			if err == nil {
				c.Dir = wd
			}
		}
		if vargs.RootDir != "" {
			c.Dir = c.Dir + "/" + vargs.RootDir
		}
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if !vargs.Sensitive {
			trace(c)
		}

		err := c.Run()
		if err != nil {
			fmt.Println("Error!")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Command completed successfully")
	}

}

func installCaCert(cacert string) *exec.Cmd {
	ioutil.WriteFile("/usr/local/share/ca-certificates/ca_cert.crt", []byte(cacert), 0644)
	return exec.Command(
		"update-ca-certificates",
	)
}

func deleteCache() *exec.Cmd {
	return exec.Command(
		"rm",
		"-rf",
		".terraform",
	)
}

func remoteConfigCommand(config remote) *exec.Cmd {
	args := []string{
		"remote",
		"config",
		fmt.Sprintf("-backend=%s", config.Backend),
	}
	for k, v := range config.Config {
		args = append(args, fmt.Sprintf("-backend-config=%s=%s", k, v))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func getModules() *exec.Cmd {
	return exec.Command(
		"terraform",
		"get",
	)
}

func planCommand(variables map[string]string, parallelism int) *exec.Cmd {
	args := []string{
		"plan",
		"-out=plan.tfout",
	}
	for k, v := range variables {
		args = append(args, "-var")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	if parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", parallelism))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func applyCommand(parallelism int) *exec.Cmd {
	args := []string{
		"apply",
	}
	if parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", parallelism))
	}
	args = append(args, "plan.tfout")
	return exec.Command(
		"terraform",
		args...,
	)
}

func assumeRole(roleArn string) {
	client := sts.New(session.New())
	duration := time.Hour * 1
	stsProvider := &stscreds.AssumeRoleProvider{
		Client:          client,
		Duration:        duration,
		RoleARN:         roleArn,
		RoleSessionName: "drone",
	}

	value, err := credentials.NewCredentials(stsProvider).Get()
	if err != nil {
		fmt.Println("Error assuming role!")
		fmt.Println(err)
		os.Exit(1)
	}
	os.Setenv("AWS_ACCESS_KEY_ID", value.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", value.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", value.SessionToken)
}

func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
