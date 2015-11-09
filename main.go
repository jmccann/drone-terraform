package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/drone/drone-plugin-go/plugin"
)

type Terraform struct {
	Commands []string `json:"commands"`
}

func main() {

	workspace := plugin.Workspace{}
	vargs := Terraform{}

	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	//skip if no commands are specified
	if len(vargs.Commands) == 0 {
		return
	}

	for _, c := range vargs.Commands {
		cmd := command(c)
		cmd.Env = os.Environ()
		cmd.Dir = workspace.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}

}

func command(cmd string) *exec.Cmd {
	args := strings.Split(cmd, " ")
	return exec.Command(args[0], args[1:]...)
}

func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
