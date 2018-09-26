# drone-terraform

[![Build Status](http://beta.drone.io/api/badges/quay.io/agari/agari-drone-terraform/status.svg)](http://beta.drone.io/quay.io/agari/agari-drone-terraform)

Drone plugin to execute Terraform plan and apply. For the usage information and
a listing of the available options please take a look at [the docs](https://github.com/quay.io/agari/agari-drone-terraform/blob/master/DOCS.md).

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo
docker build --rm=true -t quay.io/agari/agari-drone-terraform .
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-terraform' not found or does not exist.
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  quay.io/agari/agari-drone-terraform:latest --plan
```

## Drone 0.4

Legacy `drone-terraform` plugin exists @ `quay.io/agari/agari-drone-terraform:0.4`
