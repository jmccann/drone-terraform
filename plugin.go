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
		Planfile    string
		Difffile    string
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

	TfCommand struct {
		Tfcmd *exec.Cmd
		Ofile string
	}
)

// Exec executes the plugin
func (p Plugin) Exec() error {

	// Install a extra PEM key if required
	if len(os.Getenv("PEM_NAME")) > 0 {
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
		sshconfErr := installGithubSsh(os.Getenv("GITHUB_PRIVATE_SSH_KEY"))

		if sshconfErr != nil {
			return sshconfErr
		}
	}

	// Install an AWS profile if env var is set
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

	var commands []TfCommand

	commands = append(commands, TfCommand{Tfcmd: exec.Command("terraform", "version")})

	CopyTfEnv()

	if p.Config.Cacert != "" {
		commands = append(commands, TfCommand{Tfcmd: installCaCert(p.Config.Cacert)})
	}

	commands = append(commands, TfCommand{Tfcmd: deleteCache()})
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
		case "show":
			commands = append(commands, tfShow(p.Config))
		case "destroy":
			commands = append(commands, tfDestroy(p.Config))
		default:
			return fmt.Errorf("valid actions are: validate, plan, show, apply, plan-destroy, destroy.  You provided %s", action)
		}
	}

	commands = append(commands, TfCommand{Tfcmd: deleteCache()})

	for _, c := range commands {
		if c.Tfcmd.Dir == "" {
			wd, err := os.Getwd()
			if err == nil {
				c.Tfcmd.Dir = wd
			}
		}
		if p.Config.RootDir != "" {
			c.Tfcmd.Dir = c.Tfcmd.Dir + "/" + p.Config.RootDir
		}

		if c.Ofile == "" {
			c.Tfcmd.Stdout = os.Stdout
			c.Tfcmd.Stderr = os.Stderr
			if !p.Config.Sensitive {
				trace(c.Tfcmd)
			}

			err := c.Tfcmd.Run()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to execute a command")
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"file":    c.Ofile,
				"command": strings.Join(c.Tfcmd.Args, " "),
			}).Info("Command")

			out, err := c.Tfcmd.CombinedOutput()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"command": strings.Join(c.Tfcmd.Args, " "),
					"error":   err,
				}).Fatal("Failed to execute a command")
			}
			f, outferr := os.Create(c.Ofile)
			if outferr != nil {
				logrus.WithFields(logrus.Fields{
					"command": strings.Join(c.Tfcmd.Args, " "),
					"error":   outferr,
				}).Fatal("Failed to write file")
			}
			f.Write(out)
			f.Sync()
			logrus.WithFields(logrus.Fields{
				"file":      c.Ofile,
				"contenets": string(out),
			}).Info("Logging output")
			f.Close()

		}
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

func getModules() TfCommand {
	cmd := exec.Command(
		"terraform",
		"get",
	)
	return (TfCommand{Tfcmd: cmd})
}

func initCommand(config InitOptions) TfCommand {
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

	cmd := exec.Command(
		"terraform",
		args...,
	)
	return (TfCommand{Tfcmd: cmd})
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

func tfShow(config Config) TfCommand {
	args := []string{
		"show",
		"-no-color",
	}
	ofile := config.Difffile
	if config.Difffile != "" {
		args = append(args, config.Planfile)
	}

	cmd := exec.Command(
		"terraform",
		args...,
	)

	return (TfCommand{Tfcmd: cmd, Ofile: ofile})

}

func tfApply(config Config) TfCommand {
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
	if config.Planfile != "" {
		args = append(args, config.Planfile)
	} else {
		args = append(args, "plan.tfout")
	}
	cmd := exec.Command(
		"terraform",
		args...,
	)
	return (TfCommand{Tfcmd: cmd})
}

func tfDestroy(config Config) TfCommand {
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
	cmd := exec.Command(
		"terraform",
		args...,
	)
	return (TfCommand{Tfcmd: cmd})
}

func tfPlan(config Config, destroy bool) TfCommand {
	args := []string{
		"plan",
	}

	logrus.WithFields(logrus.Fields{
		"Config.Parallelism": config.Parallelism,
		"Config.Planfile":    config.Planfile,
	}).Info("Configuration")

	if destroy {
		args = append(args, "-destroy")
	} else if config.Planfile != "" {
		args = append(args, fmt.Sprintf("-out=%s", config.Planfile))
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
	cmd := exec.Command(
		"terraform",
		args...,
	)
	return (TfCommand{Tfcmd: cmd})
}

func tfValidate(config Config) TfCommand {
	args := []string{
		"validate",
	}
	for _, v := range config.VarFiles {
		args = append(args, fmt.Sprintf("-var-file=%s", v))
	}
	for k, v := range config.Vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}
	cmd := exec.Command(
		"terraform",
		args...,
	)
	return (TfCommand{Tfcmd: cmd})
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
