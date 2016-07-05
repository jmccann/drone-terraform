# Docker image for the Drone Terraform plugin
#
#     cd $GOPATH/src/github.com/drone-plugins/drone-terraform
#     make deps build docker

FROM alpine:3.4

RUN echo "@testing http://dl-cdn.alpinelinux.org/alpine/edge/testing" | tee -a /etc/apk/repositories && \
  apk -U add \
    ca-certificates \
    git \
    terraform@testing && \
  rm -rf /var/cache/apk/*

ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
