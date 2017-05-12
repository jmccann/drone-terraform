package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type (
	Config struct {
		Plan        bool
		Vars        map[string]string
		Secrets     map[string]string
		InitOptions InitOptions
		Cacert      string
		Sensitive   bool
		RoleARN     string
		RootDir     string
		Parallelism int
		Targets     []string
		VarFiles    []string
	}

	InitOptions struct {
		BackendConfig []string `json:"backend-config"`
		Lock          *bool    `json:"lock"`
		LockTimeout   string   `json:"lock-timeout"`
	}

	Plugin struct {
		Config Config
	}
)

func (p Plugin) Exec() error {
	if p.Config.RoleARN != "" {
		assumeRole(p.Config.RoleARN)
	}

	var commands []*exec.Cmd

	if len(p.Config.Secrets) != 0 {
		exportSecrets(p.Config.Secrets)
	}

	if p.Config.Cacert != "" {
		commands = append(commands, installCaCert(p.Config.Cacert))
	}

	commands = append(commands, deleteCache())

	commands = append(commands, initCommand(p.Config.InitOptions))

	commands = append(commands, getModules())
	commands = append(commands, validateCommand())
	commands = append(commands, planCommand(p.Config))
	if !p.Config.Plan {
		commands = append(commands, applyCommand(p.Config))
	}
	commands = append(commands, deleteCache())

	for _, c := range commands {
		if c.Dir == "" {
			wd, err := os.Getwd()
			if err == nil {
				c.Dir = wd
			}
		}
		if p.Config.RootDir != "" {
			c.Dir = c.Dir + "/" + p.Config.RootDir
		}
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if !p.Config.Sensitive {
			trace(c)
		}

		err := c.Run()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to execute a command")
		}
		logrus.Debug("Command completed successfully")
	}

	return nil
}

func installCaCert(cacert string) *exec.Cmd {
	ioutil.WriteFile("/usr/local/share/ca-certificates/ca_cert.crt", []byte(cacert), 0644)
	return exec.Command(
		"update-ca-certificates",
	)
}

func exportSecrets(secrets map[string]string) {
	for k, v := range secrets {
		os.Setenv(fmt.Sprintf("%s", k), fmt.Sprintf("%s", os.Getenv(v)))
	}
}

func deleteCache() *exec.Cmd {
	return exec.Command(
		"rm",
		"-rf",
		".terraform",
	)
}

func initCommand(config InitOptions) *exec.Cmd {
	args := []string{
		"init",
	}

	for _, v := range config.BackendConfig {
		args = append(args, fmt.Sprintf("-backend-config=%s", v))
	}

	// True is default in TF
	if config.Lock != nil {
		args = append(args, fmt.Sprintf("-lock=%t", *config.Lock))
	}

	// "0s" is default in TF
	if config.LockTimeout != "" {
		args = append(args, fmt.Sprintf("-lock-timeout=%s", config.LockTimeout))
	}

	// Fail Terraform execution on prompt
	args = append(args, "-input=false")

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

func validateCommand() *exec.Cmd {
	args := []string{
		"validate",
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func planCommand(config Config) *exec.Cmd {
	args := []string{
		"plan",
		"-out=plan.tfout",
	}
	for _, v := range config.Targets {
		args = append(args, "--target", fmt.Sprintf("%s", v))
	}
	for _, v := range config.VarFiles {
		args = append(args, "-var-file", fmt.Sprintf("%s", v))
	}
	for k, v := range config.Vars {
		args = append(args, "-var")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range config.Secrets {
		args = append(args, "-var")
		args = append(args, fmt.Sprintf("%s=%s", k, os.Getenv(v)))
	}
	if config.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", config.Parallelism))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func applyCommand(config Config) *exec.Cmd {
	args := []string{
		"apply",
	}
	for _, v := range config.Targets {
		args = append(args, "--target", fmt.Sprintf("%s", v))
	}
	if config.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", config.Parallelism))
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
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Error assuming role!")
	}
	os.Setenv("AWS_ACCESS_KEY_ID", value.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", value.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", value.SessionToken)
}

func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
