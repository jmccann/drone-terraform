# drone-terraform

[![Build Status](http://beta.drone.io/api/badges/jmccann/drone-terraform/status.svg)](http://beta.drone.io/jmccann/drone-terraform)
[![Join the chat at https://gitter.im/drone/drone](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/drone/drone)
[![Go Doc](https://godoc.org/github.com/jmccann/drone-terraform?status.svg)](http://godoc.org/github.com/jmccann/drone-terraform)
[![Go Report](https://goreportcard.com/badge/github.com/jmccann/drone-terraform)](https://goreportcard.com/report/github.com/jmccann/drone-terraform)
[![](https://images.microbadger.com/badges/image/jmccann/drone-terraform.svg)](https://microbadger.com/images/jmccann/drone-terraform "Get your own image badge on microbadger.com")

Drone plugin to execute Terraform plan and apply. For the usage information and
a listing of the available options please take a look at [the docs](DOCS.md).

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o release/linux/amd64/drone-terraform
docker build --rm=true -t jmccann/drone-terraform .
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
  jmccann/drone-terraform:1 --plan
```

## Drone 0.4

Legacy `drone-terraform` plugin exists @ `jmccann/drone-terraform:0.4`
