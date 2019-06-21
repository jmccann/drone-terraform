## 6.0-0.12.2 (2019-06-20)
* Update embedded TF to `0.12.2`

## 6.0-0.12.1 (2019-06-05)
* Version bump plugin to `6.0` since terraform `0.12` has breaking changes
* Update embedded TF to `0.12.1`

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
