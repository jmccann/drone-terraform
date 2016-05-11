# Docker image for the Drone Cloud Foundry plugin
#
#     cd $GOPATH/src/github.com/drone-plugins/drone-terraform
#     make deps build docker

FROM alpine:3.2

RUN apk update && \
  apk add ca-certificates && \
  rm -rf /var/cache/apk/*

ADD terraform /bin/
ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
