---
date: 2016-01-01T00:00:00+00:00
title: Terraform
author: jmccann
tags: [ infrastructure, build tool ]
repo: quay.io/agari/agari-drone-terraform
logo: terraform.svg
image: quay.io/agari/agari-drone-terraform
---

The Terraform plugin applies the infrastructure configuration contained within the repository. The below pipeline configuration demonstrates simple usage which will run a `validate`, `plan` and `apply`:

```yaml
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
```

Example configuration passing `vars` to terraform commands:

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   vars:
+     app_name: my-project
+     app_version: 1.0.0
```

Example of explicitly specifying `actions` to perform a dry run.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   actions:
+     - validate
+     - plan
```

Example configuration passing secrets to terraform.  Please read
https://www.terraform.io/docs/configuration/variables.html#environment-variables
for more details.

**Drone 0.6+**:

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   secrets:
+     - source: terraform_secret
+       target: tf_var_my_secret
```

**Drone 0.5**:

```diff
pipeline:
  terraform_1:
    image: quay.io/agari/agari-drone-terraform:5
+   environment:
+     TF_VAR_MY_SECRET: ${TERRAFORM_SECRET}

  terraform_2:
    image: quay.io/agari/agari-drone-terraform:5
    plan: false
+   sensitive: true
+   vars:
+     my_secret: ${TERRAFORM_SECRET}
```

You may be passing sensitive vars to your terraform commands.  If you do not want
the terraform commands to display in your drone logs then set `sensitive` to `true`.
The output from the commands themselves will still display, it just won't show
what command is actually being ran.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   sensitive: true
```

Example configuration for overriding the terraform version.  This will increase
plugin execution time as it will download/unpack the version of terraform
specified instead of using the embedded version that is included.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   tf_version: 0.10.3
```

Example configuration with state tracked via remote.  You will need a
[backend configuration](https://www.terraform.io/docs/backends/config.html)
specified in a `.tf` file.  You can then pass additional options via the `.drone.yml`.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   init_options:
+     backend-config:
+       - "bucket=my-terraform-config-bucket"
+       - "key=tf-states/my-project"
+       - "region=us-east-1"
```

You may want to run terraform against internal resources, like an internal
OpenStack deployment.  Sometimes these resources are signed by an internal
CA Certificate.  You can inject your CA Certificate into the plugin by using
`ca_certs` key as described above.  Below is an example.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   ca_cert: |
+     -----BEGIN CERTIFICATE-----
+     asdfsadf
+     asdfsadf
+     -----END CERTIFICATE-------
```

You may want to assume another role before running the terraform commands.
This is useful for cross account access, where a central account has privileges
to assume roles in other accounts. Using the current credentials, this role will
be assumed and exported to environment variables.
See [the discussion](https://github.com/hashicorp/terraform/issues/1275) in the Terraform issues.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   role_arn_to_assume: arn:aws:iam::account-of-role-to-assume:role/name-of-role
```

You may want to change directories before applying the terraform commands.
This parameter is useful if you have multiple environments in different folders
and you want to use different drone configurations to apply different environments.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   root_dir: some/path/here
```

You may want to only target a specific list of resources within your terraform
code. To achieve this you can specify the `targets` parameter. If left undefined
all resources will be planned/applied against as the default behavior.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   targets:
+     - aws_security_group.generic_sg
+     - aws_security_group.app_sg
```

You may want to limit the number of concurrent operations as Terraform walks its graph.
If you want to change Terraform's default parallelism (currently equal to 10) then set the `parallelism` parameter.

```diff
pipeline:
  terraform:
    image: quay.io/agari/agari-drone-terraform:5
+   parallelism: 2
```

Destroying the service can be done by specifying `plan-destroy` and `destroy` actions. Keep in mind that Fastly won't allow a service with active version be destroyed. Use `force_destroy` option in the service definition for terraform to handle it.

```yaml
pipeline:
  destroy:
    image: quay.io/agari/agari-drone-terraform:5
+   actions:
+     - plan-destroy
+     - destroy
```
# Environment variables

- PEM_NAME
  If this environment variable is set, a file called ~/.ssh/<PEM_NAME> will be created

- PEM_CONTENTS
  If this environment variable is set, the contents will be put in ~/.ssh/<PEM_NAME>
  Please be sure to include the "-----BEGIN RSA PRIVATE KEY-----" and end lines for a valid key

- GITHUB_PRIVATE_SSH_KEY
  If this environment variable is set, ~/.ssh/id_rsa will be set.
  Please be sure to include the "-----BEGIN RSA PRIVATE KEY-----" and end lines for a valid key

- AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
  If this environment variable is set a ~/.aws/credentials file will be added with the name of the AWS_PROFILE

- AWS_PROFILE (default: `drone-testing`)
  If a profile is created, it will be set to this name

# Parameter Reference

actions
: List of terraform actions to perform with the plugin.  List includes:
`validate`, `plan`, `apply`, `plan-destroy`, `destroy`.

init_options
: contains the configuration for the Terraform backend.

init_options.backend-config
: This specifies additional configuration to merge for the backend. This can be
specified multiple times. Flags specified later in the line override those
specified earlier if they conflict.

init_options.lock
: Lock the state file when locking is supported. Default `true`.

init_options.lock-timeout
: Duration to wait for a state lock. Default `0s`.

vars
: a map of variables to pass to the Terraform `plan` and `apply` commands.
Each value is passed as a `-var <key>=<value>` option.

var_files
: a list of variable files to pass to the Terraform `plan` and `apply` commands.
Each value is passed as a `-var-file <value>` option.

ca_cert
: ca cert to add to your environment to allow terraform to use internal/private resources

sensitive
: (default: `false`) - Whether or not to suppress terraform commands to stdout.

role_arn_to_assume
: A role to assume before running the terraform commands.

root_dir
: The root directory where the terraform files live. When unset, the top level directory will be assumed.

parallelism
: The number of concurrent operations as Terraform walks its graph.

# Testing Locally

```
docker run -e PEM_NAME=my.pem -w /root/test -v `pwd`:/root quay.io/agari/agari-drone-terraform 
```
