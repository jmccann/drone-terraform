package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type (
	// Config holds input parameters for the plugin
	Config struct {
		Actions     []string
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

	// Netrc is credentials for cloning
	Netrc struct {
		Machine  string
		Login    string
		Password string
	}

	// InitOptions include options for the Terraform's init command
	InitOptions struct {
		BackendConfig []string `json:"backend-config"`
		Lock          *bool    `json:"lock"`
		LockTimeout   string   `json:"lock-timeout"`
	}

	// Plugin represents the plugin instance to be executed
	Plugin struct {
		Config    Config
		Netrc     Netrc
		Terraform Terraform
	}
)

// Exec executes the plugin
func (p Plugin) Exec() error {

	// Install a extra PEM key if required
	if len(os.Getenv("PEM_NAME")) > 0 {
		fmt.Println("--- Setting a pem file")
		value, exists := os.LookupEnv("PEM_CONTENTS")
		if !exists {
			value = "-----BEGIN RSA PRIVATE KEY-----\n\n-----END RSA PRIVATE KEY-----\n"
		}
		err := installExtraPem(os.Getenv("PEM_NAME"), value)

		if err != nil {
			return err
		}
	}

	// Install a Github SSH key
	if len(os.Getenv("GITHUB_PRIVATE_SSH_KEY")) > 0 {
		fmt.Println("--- Setting a Github key")
		sshconfErr := installGithubSsh(os.Getenv("GITHUB_PRIVATE_SSH_KEY"))

		if sshconfErr != nil {
			return sshconfErr
		}
	}

	// Install an AWS profile if env var is set
	fmt.Println("--- Setting an AWS profile")
	if len(os.Getenv("AWS_ACCESS_KEY_ID")) > 0 {
		profileErr := installProfile(os.Getenv("AWS_PROFILE"), os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"))

		if profileErr != nil {
			return profileErr
		}
	}
	// Install specified version of terraform
	if p.Terraform.Version != "" {

		err := installTerraform(p.Terraform.Version)

		if err != nil {
			return err
		}
	}

	if p.Config.RoleARN != "" {
		assumeRole(p.Config.RoleARN)
	}

	// writing the .netrc file with Github credentials in it.
	err := writeNetrc(p.Netrc.Machine, p.Netrc.Login, p.Netrc.Password)
	if err != nil {
		return err
	}

	var commands []*exec.Cmd

	commands = append(commands, exec.Command("terraform", "version"))

	CopyTfEnv()

	if p.Config.Cacert != "" {
		commands = append(commands, installCaCert(p.Config.Cacert))
	}

	commands = append(commands, deleteCache())
	commands = append(commands, initCommand(p.Config.InitOptions))
	commands = append(commands, getModules())

	// Add commands listed from Actions
	for _, action := range p.Config.Actions {
		switch action {
		case "validate":
			commands = append(commands, tfValidate(p.Config))
		case "plan":
			commands = append(commands, tfPlan(p.Config, false))
		case "plan-destroy":
			commands = append(commands, tfPlan(p.Config, true))
		case "apply":
			commands = append(commands, tfApply(p.Config))
		case "destroy":
			commands = append(commands, tfDestroy(p.Config))
		default:
			return fmt.Errorf("valid actions are: validate, plan, apply, plan-destroy, destroy.  You provided %s", action)
		}
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

// CopyTfEnv creates copies of TF_VAR_ to lowercase
func CopyTfEnv() {
	tfVar := regexp.MustCompile(`^TF_VAR_.*$`)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if tfVar.MatchString(pair[0]) {
			name := strings.Split(pair[0], "TF_VAR_")
			os.Setenv(fmt.Sprintf("TF_VAR_%s", strings.ToLower(name[1])), pair[1])
		}
	}
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

func deleteCache() *exec.Cmd {
	return exec.Command(
		"rm",
		"-rf",
		".terraform",
	)
}

func getModules() *exec.Cmd {
	return exec.Command(
		"terraform",
		"get",
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

func installCaCert(cacert string) *exec.Cmd {
	ioutil.WriteFile("/usr/local/share/ca-certificates/ca_cert.crt", []byte(cacert), 0644)
	return exec.Command(
		"update-ca-certificates",
	)
}

func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

func tfApply(config Config) *exec.Cmd {
	args := []string{
		"apply",
	}
	for _, v := range config.Targets {
		args = append(args, "--target", fmt.Sprintf("%s", v))
	}
	if config.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", config.Parallelism))
	}
	if config.InitOptions.Lock != nil {
		args = append(args, fmt.Sprintf("-lock=%t", *config.InitOptions.Lock))
	}
	if config.InitOptions.LockTimeout != "" {
		args = append(args, fmt.Sprintf("-lock-timeout=%s", config.InitOptions.LockTimeout))
	}
	args = append(args, "plan.tfout")
	return exec.Command(
		"terraform",
		args...,
	)
}

func tfDestroy(config Config) *exec.Cmd {
	args := []string{
		"destroy",
	}
	for _, v := range config.Targets {
		args = append(args, fmt.Sprintf("-target=%s", v))
	}
	args = append(args, varFiles(config.VarFiles)...)
	args = append(args, vars(config.Vars)...)
	if config.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", config.Parallelism))
	}
	if config.InitOptions.Lock != nil {
		args = append(args, fmt.Sprintf("-lock=%t", *config.InitOptions.Lock))
	}
	if config.InitOptions.LockTimeout != "" {
		args = append(args, fmt.Sprintf("-lock-timeout=%s", config.InitOptions.LockTimeout))
	}
	args = append(args, "-force")
	return exec.Command(
		"terraform",
		args...,
	)
}

func tfPlan(config Config, destroy bool) *exec.Cmd {
	args := []string{
		"plan",
	}

	if destroy {
		args = append(args, "-destroy")
	} else {
		args = append(args, "-out=plan.tfout")
	}

	for _, v := range config.Targets {
		args = append(args, "--target", fmt.Sprintf("%s", v))
	}
	args = append(args, varFiles(config.VarFiles)...)
	args = append(args, vars(config.Vars)...)
	if config.Parallelism > 0 {
		args = append(args, fmt.Sprintf("-parallelism=%d", config.Parallelism))
	}
	if config.InitOptions.Lock != nil {
		args = append(args, fmt.Sprintf("-lock=%t", *config.InitOptions.Lock))
	}
	if config.InitOptions.LockTimeout != "" {
		args = append(args, fmt.Sprintf("-lock-timeout=%s", config.InitOptions.LockTimeout))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func tfValidate(config Config) *exec.Cmd {
	args := []string{
		"validate",
	}
	for _, v := range config.VarFiles {
		args = append(args, fmt.Sprintf("-var-file=%s", v))
	}
	for k, v := range config.Vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}
	return exec.Command(
		"terraform",
		args...,
	)
}

func vars(vs map[string]string) []string {
	var args []string
	for k, v := range vs {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

func varFiles(vfs []string) []string {
	var args []string
	for _, v := range vfs {
		args = append(args, fmt.Sprintf("-var-file=%s", v))
	}
	return args
}

// helper function to write a netrc file.
// The following code comes from the official Git plugin for Drone:
// https://github.com/drone-plugins/drone-git/blob/8386effd2fe8c8695cf979427f8e1762bd805192/utils.go#L43-L68
func writeNetrc(machine, login, password string) error {
	if machine == "" {
		return nil
	}
	out := fmt.Sprintf(
		netrcFile,
		machine,
		login,
		password,
	)

	home := "/root"
	u, err := user.Current()
	if err == nil {
		home = u.HomeDir
	}
	path := filepath.Join(home, ".netrc")
	return ioutil.WriteFile(path, []byte(out), 0600)
}

const netrcFile = `
machine %s
login %s
password %s
`
