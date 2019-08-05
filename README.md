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

## Docker

Build the docker image with the following commands:

```
docker build --rm=true \
  -t jmccann/drone-terraform \
  --build-arg terraform_version=0.12.0 .
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  jmccann/drone-terraform:latest --plan
```

## Drone 0.4

Legacy `drone-terraform` plugin exists @ `jmccann/drone-terraform:0.4`
