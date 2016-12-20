Use the Terraform plugin to apply the infrastructure configuration contained within the repository. The following parameters are used to configure this plugin:

* `plan` - if true, calculates a plan but does __NOT__ apply it.
* `remote` - contains the configuration for the Terraform remote state tracking.
  * `backend` - the Terraform remote state backend to use.
  * `config` - a map of configuration parameters for the remote state backend. Each value is passed as a `-backend-config=<key>=<value>` option.
* `vars` - a map of variables to pass to the Terraform `plan` and `apply` commands. Each value is passed as a `-var
 <key>=<value>` option.
* `secrets` - a map of variables to pass to the Terraform `plan` and `apply` commands.  Each value is passed as a `-var
 <key>=<ENVVAR>` option.  The `ENVVAR` is read as the key/pair value.
* `ca_cert` - ca cert to add to your environment to allow terraform to use internal/private resources
* `sensitive` (default: `false`) - Whether or not to suppress terraform commands to stdout.
* `role_arn_to_assume` - A role to assume before running the terraform commands.
* `root_dir` - The root directory where the terraform files live. When unset, the top level directory will be assumed.
* `parallelism` - The number of concurrent operations as Terraform walks its graph.

The following is a sample Terraform configuration in your .drone.yml file:

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
    vars:
      app_name: my-project
      app_version: 1.0.0
    secrets:
      my_secret: TERRAFORM_SECRET
```

# Advanced Configuration

## CA Certs
You may want to run terraform against internal resources, like an internal
OpenStack deployment.  Usually these resources are signed by an internal
CA Certificate.  You can inject your CA Certificate into the plugin by using
`ca_certs` key as described above.  Below is an example.

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    remote:
      backend: swift
      config:
        path: drone/terraform
    vars:
      app_name: my-project
      app_version: 1.0.0
    ca_cert: |
      -----BEGIN CERTIFICATE-----
      asdfsadf
      asdfsadf
      -----END CERTIFICATE-------
```

## Suppress Sensitive Output
You may be passing sensitive vars to your terraform commands.  If you do not want
the terraform commands to display in your drone logs then set `sensitive` to `true`.
The output from the commands themselves will still display, it just won't show
want command is actually being ran.

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    sensitive: true
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
      vars:
        app_name: my-project
        app_version: 1.0.0
```

## Assume Role ARN
You may want to assume another role before running the terraform commands. This is useful for cross account access, where a central account ahs privileges to assume roles in other accounts. Using the current credentials, this role will be assumed and exported to environment variables.  See [the discussion](https://github.com/hashicorp/terraform/issues/1275) in the Terraform issues.

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
    vars:
      app_name: my-project
      app_version: 1.0.0
    role_arn_to_assume: arn:aws:iam::account-of-role-to-assume:role/name-of-role
```

## Root dir
You may want to change directories before applying the terraform commands.  This parameter is useful if you have multiple environments in different folders and you want to use different drone configurations to apply different environments.

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
    vars:
      app_name: my-project
      app_version: 1.0.0
    root_dir: some/path/here
```

## Targets
You may want to only target a specific list of resources within your terraform code. To achieve this you can specify the `targets` parameter. If left undefined all resources will be planned/applied against as the default behavior.

Single target:

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    targets: aws_security_group.generic_sg
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
    vars:
      app_name: my-project
      app_version: 1.0.0
```

Multiple targets:

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    targets:
      - aws_security_group.generic_sg
      - aws_security_group.app_sg
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
    vars:
      app_name: my-project
      app_version: 1.0.0
```

## Parallelism
You may want to limit the number of concurrent operations as Terraform walks its graph.
If you want to change Terraform's default parallelism (currently equal to 10) then set the `parallelism` parameter.

```yaml
pipeline:
  terraform:
    image: jmccann/drone-terraform:0.5
    plan: false
    remote:
      backend: S3
      config:
        bucket: my-terraform-config-bucket
        key: tf-states/my-project
        region: us-east-1
    vars:
      app_name: my-project
      app_version: 1.0.0
    parallelism: 2
```
