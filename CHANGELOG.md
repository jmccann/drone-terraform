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
