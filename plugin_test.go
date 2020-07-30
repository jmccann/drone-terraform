package main

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/franela/goblin"
)

func TestPlugin(t *testing.T) {
	os.Setenv("TEST_USER", "johndoe")
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

	g.Describe("credsSet", func() {
		tests := []struct {
			name string
			args map[string]string
			want bool
		}{
			{
				"Should return true when all credentials were set",
				map[string]string{"AWS_ACCESS_KEY_ID": "x", "AWS_SECRET_ACCESS_KEY": "x", "AWS_SESSION_TOKEN": "x"},
				true,
			},
			{
				"Should return false when access key id is missing",
				map[string]string{"AWS_SECRET_ACCESS_KEY": "x", "AWS_SESSION_TOKEN": "x"},
				false,
			},
			{
				"Should return false when secret access key is missing",
				map[string]string{"AWS_ACCESS_KEY_ID": "x", "AWS_SESSION_TOKEN": "x"},
				false,
			},
			{
				"Should return false when session token is missing",
				map[string]string{"AWS_ACCESS_KEY_ID": "x", "AWS_SECRET_ACCESS_KEY": "x"},
				false,
			},
		}

		for _, tt := range tests {
			g.It(tt.name, func() {
				for k, v := range tt.args {
					os.Setenv(k, v)
				}
				g.Assert(credsSet()).Equal(tt.want)
			})
		}
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
					"with parallelism",
					args{config: Config{Parallelism: 5}},
					exec.Command("terraform", "apply", "-parallelism=5", "plan.tfout"),
				},
				{
					"with targets",
					args{config: Config{Targets: []string{"target1", "target2"}}},
					exec.Command("terraform", "apply", "--target", "target1", "--target", "target2", "plan.tfout"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfApply(tt.args.config).Args).Equal(tt.want.Args)
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
					"with parallelism",
					args{config: Config{Parallelism: 5}},
					exec.Command("terraform", "destroy", "-parallelism=5", "-force"),
				},
				{
					"with targets",
					args{config: Config{Targets: []string{"target1", "target2"}}},
					exec.Command("terraform", "destroy", "-target=target1", "-target=target2", "-force"),
				},
				{
					"with vars",
					args{config: Config{Vars: map[string]string{"username": "someuser", "password": "1pass"}}},
					exec.Command("terraform", "destroy", "-var", "username=someuser", "-var", "password=1pass", "-force"),
				},
				{
					"with vars as env vars",
					args{config: Config{Vars: map[string]string{"testuser": "${TEST_USER}"}}},
					exec.Command("terraform", "destroy", "-var", "testuser=johndoe", "-force"),
				},
				{
					"with var-files",
					args{config: Config{VarFiles: []string{"common.tfvars", "prod.tfvars"}}},
					exec.Command("terraform", "destroy", "-var-file=common.tfvars", "-var-file=prod.tfvars", "-force"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfDestroy(tt.args.config)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfValidate", func() {
		g.It("Should return correct validate command", func() {
			tests := []struct {
				name string
				want *exec.Cmd
			}{
				{
					"default",
					exec.Command("terraform", "validate"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfValidate()).Equal(tt.want)
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
				{
					"with vars",
					args{config: Config{Vars: map[string]string{"username": "someuser", "password": "1pass"}}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout", "-var", "username=someuser", "-var", "password=1pass"),
				},
				{
					"with vars as env vars",
					args{config: Config{Vars: map[string]string{"testuser": "${TEST_USER}"}}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout", "-var", "testuser=johndoe"),
				},
				{
					"with var-files",
					args{config: Config{VarFiles: []string{"common.tfvars", "prod.tfvars"}}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout", "-var-file=common.tfvars", "-var-file=prod.tfvars"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfPlan(tt.args.config, tt.destroy)).Equal(tt.want)
			}
		})
	})
	g.Describe("tfFmt", func() {
		g.It("Should return correct fmt commands given the arguments", func() {
			type args struct {
				config Config
			}

			affirmative := true
			negative := false

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					exec.Command("terraform", "fmt"),
				},
				{
					"with list",
					args{config: Config{FmtOptions: FmtOptions{List: &affirmative}}},
					exec.Command("terraform", "fmt", "-list=true"),
				},
				{
					"with write",
					args{config: Config{FmtOptions: FmtOptions{Write: &affirmative}}},
					exec.Command("terraform", "fmt", "-write=true"),
				},
				{
					"with diff",
					args{config: Config{FmtOptions: FmtOptions{Diff: &affirmative}}},
					exec.Command("terraform", "fmt", "-diff=true"),
				},
				{
					"with check",
					args{config: Config{FmtOptions: FmtOptions{Check: &affirmative}}},
					exec.Command("terraform", "fmt", "-check=true"),
				},
				{
					"with combination",
					args{config: Config{FmtOptions: FmtOptions{
						List:  &negative,
						Write: &negative,
						Diff:  &affirmative,
						Check: &affirmative,
					}}},
					exec.Command("terraform", "fmt", "-list=false", "-write=false", "-diff=true", "-check=true"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfFmt(tt.args.config)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfDataDir", func() {
		g.It("Should override the terraform data dir environment variable when provided", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"with TerraformDataDir",
					args{config: Config{TerraformDataDir: ".overriden_terraform_dir"}},
					exec.Command("terraform", "apply", ".overriden_terraform_dir.plan.tfout"),
				},
				{
					"with TerraformDataDir value as .terraform",
					args{config: Config{TerraformDataDir: ".terraform"}},
					exec.Command("terraform", "apply", "plan.tfout"),
				},
				{
					"without TerraformDataDir",
					args{config: Config{}},
					exec.Command("terraform", "apply", "plan.tfout"),
				},
			}

			for _, tt := range tests {
				os.Setenv("TF_DATA_DIR", tt.args.config.TerraformDataDir)
				applied := tfApply(tt.args.config)

				g.Assert(applied).Equal(tt.want)

			}
		})
	})
}
