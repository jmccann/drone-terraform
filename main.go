package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/drone/drone-plugin-go/plugin"
)

var (
	buildCommit string
)

type terraform struct {
	Remote    remote            `json:"remote"`
	Plan      bool              `json:"plan"`
	Vars      map[string]string `json:"vars"`
	Cacert    string            `json:"ca_cert"`
	Sensitive bool              `json:"sensitive"`
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

	var commands []*exec.Cmd
	remote := vargs.Remote
	if vargs.Cacert != "" {
		commands = append(commands, installCaCert(vargs.Cacert))
	}
	if remote.Backend != "" {
		commands = append(commands, deleteCache())
		commands = append(commands, remoteConfigCommand(remote))
	}
	commands = append(commands, planCommand(vargs.Vars))
	if !vargs.Plan {
		commands = append(commands, applyCommand())
	}
	commands = append(commands, deleteCache())

	for _, c := range commands {
		c.Env = os.Environ()
		c.Dir = workspace.Path
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

func planCommand(variables map[string]string) *exec.Cmd {
	args := []string{
		"plan",
		"-out=plan.tfout",
	}
	for k, v := range variables {
		args = append(args, "-var")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func applyCommand() *exec.Cmd {
	return exec.Command(
		"terraform",
		"apply",
		"plan.tfout",
	)
}

func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
