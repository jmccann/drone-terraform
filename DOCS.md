Use the Terraform plugin to apply the infrastructure configuration contained within the repository. The following parameters are used to configure this plugin:

* `plan` - if true, calculates a plan but does __NOT__ apply it.
* `remote` - contains the configuration for the Terraform remote state tracking.
  * `backend` - the Terraform remote state backend to use.
  * `config` - a map of configuration parameters for the remote state backend. Each value is passed as a `-backend-config=<key>=<value>` option.
* `vars` - a map of variables to pass to the Terraform `plan` and `apply` commands. Each value is passed as a `-var
 <key>=<value>` option.
* `ca_cert` - ca cert to add to your environment to allow terraform to use internal/private resources

The following is a sample Terraform configuration in your .drone.yml file:

```yaml
deploy:
  terraform:
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
```

# Advanced Configuration

## CA Certs
You may want to run terraform against internal resources, like an internal
OpenStack deployment.  Usually these resources are signed by an internal
CA Certificate.  You can inject your CA Certificate into the plugin by using
`ca_certs` key as described above.  Below is an example.

```yaml
deploy:
  terraform:
    dry_run: false
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
      -----END CERTIFICATE-----
```
