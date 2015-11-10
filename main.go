package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/drone/drone-plugin-go/plugin"
)

type terraform struct {
	Remote remote            `json:"remote"`
	DryRun bool              `json:"dryRun"`
	Vars   map[string]string `json:"vars"`
}

type remote struct {
	Backend string            `json:"backend"`
	Config  map[string]string `json:"config"`
}

func main() {

	workspace := plugin.Workspace{}
	vargs := terraform{}

	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	var commands []*exec.Cmd
	remote := vargs.Remote
	if remote.Backend != "" {
		commands = append(commands, remoteConfigCommand(remote))
	}
	commands = append(commands, planCommand(vargs.Vars))
	if !vargs.DryRun {
		commands = append(commands, applyCommand())
	}

	for _, c := range commands {
		c.Env = os.Environ()
		c.Dir = workspace.Path
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		trace(c)

		err := c.Run()
		if err != nil {
			fmt.Println("Error!")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Command completed successfully")
	}

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
