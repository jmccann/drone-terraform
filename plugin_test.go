package main

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/franela/goblin"
)

func TestPlugin(t *testing.T) {
	g := Goblin(t)

	g.Describe("CopyTfEnv", func() {
		g.It("Should create copies of TF_VAR_ to lowercase", func() {
			// Set some initial TF_VAR_ that are uppercase
			os.Setenv("TF_VAR_SOMETHING", "some value")
			os.Setenv("TF_VAR_SOMETHING_ELSE", "some other value")
			os.Setenv("TF_VAR_BASE64", "dGVzdA==")

			CopyTfEnv()

			// Make sure new env vars exist with proper values
			g.Assert(os.Getenv("TF_VAR_something")).Equal("some value")
			g.Assert(os.Getenv("TF_VAR_something_else")).Equal("some other value")
			g.Assert(os.Getenv("TF_VAR_base64")).Equal("dGVzdA==")
		})
	})

	g.Describe("tfApply", func() {
		g.It("Should return correct apply commands given the arguments", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					exec.Command("terraform", "apply", "plan.tfout"),
				},
				{
					"with targets",
					args{config: Config{Targets: []string{"target1", "target2"}}},
					exec.Command("terraform", "apply", "--target", "target1", "--target", "target2", "plan.tfout"),
				},
				{
					"with parallelism",
					args{config: Config{Parallelism: 5}},
					exec.Command("terraform", "apply", "-parallelism=5", "plan.tfout"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfApply(tt.args.config)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfDestroy", func() {
		g.It("Should return correct destroy commands given the arguments", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					exec.Command("terraform", "destroy", "-force"),
				},
				{
					"with targets",
					args{config: Config{Targets: []string{"target1", "target2"}}},
					exec.Command("terraform", "destroy", "-target=target1", "-target=target2", "-force"),
				},
				{
					"with parallelism",
					args{config: Config{Parallelism: 5}},
					exec.Command("terraform", "destroy", "-parallelism=5", "-force"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfDestroy(tt.args.config)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfPlan", func() {
		g.It("Should return correct plan commands given the arguments", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name    string
				args    args
				destroy bool
				want    *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout"),
				},
				{
					"destroy",
					args{config: Config{}},
					true,
					exec.Command("terraform", "plan", "-destroy"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfPlan(tt.args.config, tt.destroy)).Equal(tt.want)
			}
		})
	})
}
