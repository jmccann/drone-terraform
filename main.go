package main

import (
	"encoding/json"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var revision string // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "terraform plugin"
	app.Usage = "terraform plugin"
	app.Action = run
	app.Version = revision
	app.Flags = []cli.Flag{

		//
		// plugin args
		//

		cli.BoolFlag{
			Name:   "plan",
			Usage:  "calculates a plan but does NOT apply it",
			EnvVar: "PLUGIN_PLAN",
		},
		cli.StringFlag{
			Name:   "remote",
			Usage:  "contains the configuration for the Terraform remote state tracking",
			EnvVar: "PLUGIN_REMOTE",
		},
		cli.StringFlag{
			Name:   "vars",
			Usage:  "a map of variables to pass to the Terraform `plan` and `apply` commands. Each value is passed as a `<key>=<value>` option",
			EnvVar: "PLUGIN_VARS",
		},
		cli.StringFlag{
			Name:   "secrets",
			Usage:  "a map of secrets to pass to the Terraform `plan` and `apply` commands. Each value is passed as a `<key>=<ENV>` option",
			EnvVar: "PLUGIN_SECRETS",
		},
		cli.StringFlag{
			Name:   "ca_cert",
			Usage:  "ca cert to add to your environment to allow terraform to use internal/private resources",
			EnvVar: "PLUGIN_CA_CERT",
		},
		cli.BoolFlag{
			Name:   "sensitive",
			Usage:  "whether or not to suppress terraform commands to stdout",
			EnvVar: "PLUGIN_SENSITIVE",
		},
		cli.StringFlag{
			Name:   "role_arn_to_assume",
			Usage:  "A role to assume before running the terraform commands",
			EnvVar: "PLUGIN_ROLE_ARN_TO_ASSUME",
		},
		cli.StringFlag{
			Name:   "root_dir",
			Usage:  "The root directory where the terraform files live. When unset, the top level directory will be assumed",
			EnvVar: "PLUGIN_ROOT_DIR",
		},
		cli.IntFlag{
			Name:   "parallelism",
			Usage:  "The number of concurrent operations as Terraform walks its graph",
			EnvVar: "PLUGIN_PARALLELISM",
		},

		cli.StringFlag{
			Name:  "env-file",
			Usage: "source env file",
		},

		cli.StringSliceFlag{
			Name:   "targets",
			Usage:  "targets to run apply or plan on",
			EnvVar: "PLUGIN_TARGETS",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.WithFields(logrus.Fields{
		"Revision": revision,
	}).Info("Drone Terraform Plugin Version")

	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	remote := Remote{}
	json.Unmarshal([]byte(c.String("remote")), &remote)

	var vars map[string]string
	if c.String("vars") != "" {
		if err := json.Unmarshal([]byte(c.String("vars")), &vars); err != nil {
			panic(err)
		}
	}
	var secrets map[string]string
	if c.String("secrets") != "" {
		if err := json.Unmarshal([]byte(c.String("secrets")), &secrets); err != nil {
			panic(err)
		}
	}

	plugin := Plugin{
		Config: Config{
			Remote:      remote,
			Plan:        c.Bool("plan"),
			Vars:        vars,
			Secrets:     secrets,
			Cacert:      c.String("ca_cert"),
			Sensitive:   c.Bool("sensitive"),
			RoleARN:     c.String("role_arn_to_assume"),
			RootDir:     c.String("root_dir"),
			Parallelism: c.Int("parallelism"),
			Targets:     c.StringSlice("targets"),
		},
	}

	return plugin.Exec()
}
