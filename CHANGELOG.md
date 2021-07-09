## 8.3-1.0.2 (2021-07-09)
* Update embedded TF to `1.0.2`
  * A continuation of `v0.15` release line but now following semver
  * Please refer to [Terraform v1.0 Compatibility Promises](https://www.terraform.io/docs/language/v1-compatibility-promises.html)

## 8.3-0.15.1 (2021-07-09)
* fix: remove -lock and -lock-timeout for terraform init - https://github.com/jmccann/drone-terraform/pull/130

## 8.2-0.15.1 (2021-06-02)
* fix: `-force` -> `-auto-approve` - https://github.com/jmccann/drone-terraform/pull/128

## 8.1-0.15.1 (2021-04-27)
* added ability to [disable refresh](https://github.com/jmccann/drone-terraform/pull/120)

## 8.0-0.15.1 (2021-04-27)
* Update embedded TF to `0.15.1`
  * Please refer to [Terraform v15 Migration Guide](https://www.terraform.io/upgrade-guides/0-15.html)
* Update other deps

## 7.0-0.14.11 (2021-04-27)
* Update embedded TF to `0.14.11`
  * Please refer to [Terraform v14 Migration Guide](https://www.terraform.io/upgrade-guides/0-14.html)

## 6.4-0.13.2 (2020-09-11)
* Update embedded TF to `0.13.2`

## 6.4-0.13.1 (2020-09-11)
* minor version bump due to major version change of terraform
* Update embedded TF to `0.13.1`
* added curl

## 6.3-0.12.20 (2020-02-12)
* add ability to load creds from `env_file` parameter (https://github.com/jmccann/drone-terraform/pull/107).  Thanks @neemiasjnr!

## 6.2-0.12.20 (2020-02-05)
* Update embedded TF to `0.12.20`
* Update alpine to 3.11 in docker image (https://github.com/jmccann/drone-terraform/pull/109). Thanks @sgerrand!

## 6.2-0.12.16 (2019-11-26)
* Update embedded TF to `0.12.16`

## 6.2-0.12.11 (2019-11-26)
* tfValidate vars and var-file argument removal #106 Thanks @gsingh1

## 6.1-0.12.11 (2019-10-18)
* Update embedded TF to `0.12.11`

## 6.1-0.12.10 (2019-10-15)
* Update embedded TF to `0.12.10`

## 6.1-0.12.8 (2019-09-06)
* Support for parallel execution (https://github.com/jmccann/drone-terraform/pull/94).  Thanks @caioquirino!
* Updated golang to 1.13

## 6.0-0.12.8 (2019-09-06)
* Update embedded TF to `0.12.8`

## 6.0-0.12.6 (2019-08-05)
* Update embedded TF to `0.12.6`

## 6.0-0.12.4 (2019-07-12)
* Update embedded TF to `0.12.4`

## 6.0-0.12.1 (2019-06-05)
* Version bump plugin to `6.0` since terraform `0.12` has breaking changes
* Update embedded TF to `0.12.1`

## 5.3-0.11.14 (2019-08-05)
* Update embedded TF to `0.11.14`

## 5.2-0.11.13 (2019-03-27)
* Update embedded TF to `0.11.13`

## 5.2-0.11.11 (2019-02-22)
* Add `fmt` action

## 5.1-0.11.11 (2019-01-18)
* Update embedded TF to `0.11.11`

## 5.1-0.11.8 (2018-10-11)
* Update embedded TF to `0.11.8`

## 5.1-0.11.7 (2018-07-31)
* Add `vars` and `var_files` to destroy operation

## 5.0-0.11.7 (2018-04-25)
**BREAKING CHANGE**
* Removed `destroy` param
* Removed `plan` param
* Added `actions` param to provide a list of actions to perform.
See [DOCS.md](DOCS.md) for more info and examples.

## 4.1-0.11.7 (2018-04-25)
* Add .netrc support
* Update embedded TF to `0.11.7`

## 4.0-0.11.3 (2018-02-07)
* Update embedded TF to `0.11.3`

## 4.0-0.10.8 (2018-02-07)
* Pass `-var-file` to validate command
* Update embedded TF to `0.10.8`

## 4.0-0.10.7 (2017-10-20)
* Persist state locking config (https://github.com/jmccann/drone-terraform/pull/55)
* Update embedded TF to `0.10.7`

## 4.0-0.10.3 (2017-09-06)
**Breaking Change**
* Update embedded TF to 0.10.3
* In order to support validate in TF 0.10.3 add `vars` to validate command.
This is not compatible with older versions of TF.

## 3.0-0.9.11 (2017-09-06)
**Breaking Change**
* Removed `secrets` key

**Added Features**
* Added support for `destroy`
* Add ability to specify TF version to use via `tf_version`
