# drone-terraform

[![Build Status](http://beta.drone.io/api/badges/jmccann/drone-terraform/status.svg)](http://beta.drone.io/jmccann/drone-terraform)

Drone plugin to execute Terraform plan and apply. For the usage information and
a listing of the available options please take a look at [the docs](https://github.com/jmccann/drone-terraform/blob/master/DOCS.md).

## Build

Build the binary with the following commands:

```
export GO111MODULE=on
go mod download
go test
go build
```

## Usage

Execute from the working directory:

```bash
$ PLUGIN_ACTIONS=plan-fail-on-modification ./drone-terraform 
```
