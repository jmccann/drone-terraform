# drone-terraform

[![Build Status](https://drone.nwk.io/api/badges/jonathanio/drone-terraform/status.svg)](https://drone.nwk.io/jonathanio/drone-terraform)

Drone plugin to execute Terraform. For the usage information and a listing of
the available options please take a look at [the
docs](https://github.com/jonathanio/drone-terraform/blob/master/DOCS.md).

## Build

Build the binary with the following commands:

```bash
go build
go test
```

## Docker

Build the docker image with the following commands:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo
docker build --rm=true -t jonathanio/drone-terraform .
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```text
docker: Error response from daemon: Container command
'/bin/drone-terraform' not found or does not exist.
```

## Usage

Execute from the working directory:

```bash
docker run --rm \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  jonathanio/drone-terraform:latest --plan
```

## Legacy and Upstream

Access to older versions and upstream versions of `drone-terraform` is
available from the master repository at
[jmccann/drone-terraform](https://github.com/jmccann/drone-terraform).
