package main

import (
	"encoding/hex"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

type (
	Config struct {
		Remote      Remote
		Plan        bool
		Vars        map[string]string
		Secrets     map[string]string
		Cacert      string
		Sensitive   bool
		RoleARN     string
		RootDir     string
		Parallelism int
		Targets     []string
		Submodules  map[string]map[string]string
	}

	Remote struct {
		Backend string            `json:"backend"`
		Config  map[string]string `json:"config"`
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
	remote := p.Config.Remote

	if len(p.Config.Secrets) != 0 {
		exportSecrets(p.Config.Secrets)
	}

	if len(p.Config.Submodules) != 0 {
		submoduleOverride(p.Config.Submodules)
	}

	if p.Config.Cacert != "" {
		commands = append(commands, installCaCert(p.Config.Cacert))
	}
	if remote.Backend != "" {
		commands = append(commands, deleteCache())
		commands = append(commands, remoteConfigCommand(remote))
	}
	commands = append(commands, getModules())
	commands = append(commands, planCommand(p.Config.Vars, p.Config.Secrets, p.Config.Parallelism, p.Config.Targets))
	if !p.Config.Plan {
		commands = append(commands, applyCommand(p.Config.Parallelism, p.Config.Targets))
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

func submoduleOverride(submodules map[string]map[string]string) {
	allOverrides := []string{}
	for moduleName, override := range submodules {
		overrideContents := []string{}
		for k, v := range override {
			overrideString := fmt.Sprintf(`  %s = "%s"`, k, v)
			overrideContents = append(overrideContents, overrideString)
		}
		moduleContents := fmt.Sprintf(`
module "%s" {
%s
}`, moduleName, strings.Join(overrideContents, "\n"))
		allOverrides = append(allOverrides, moduleContents)
	}
	fileContents := []byte(strings.Join(allOverrides, "\n"))
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	fileName := hex.EncodeToString(randBytes) + "_override.tf"
	err := ioutil.WriteFile(fileName, fileContents, 0644)
	if err != nil {
		panic(err)
	}
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

func remoteConfigCommand(config Remote) *exec.Cmd {
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

func planCommand(variables map[string]string, secrets map[string]string, parallelism int, targets []string) *exec.Cmd {
	args := []string{
		"plan",
		"-out=plan.tfout",
	}
	for _, v := range targets {
		args = append(args, "--target", fmt.Sprintf("%s", v))
	}
	for k, v := range variables {
		args = append(args, "-var")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range secrets {
		args = append(args, "-var")
		args = append(args, fmt.Sprintf("%s=%s", k, os.Getenv(v)))
	}
	if parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", parallelism))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func applyCommand(parallelism int, targets []string) *exec.Cmd {
	args := []string{
		"apply",
	}
	for _, v := range targets {
		args = append(args, "--target", fmt.Sprintf("%s", v))
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
