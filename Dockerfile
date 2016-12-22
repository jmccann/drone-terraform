# Docker image for the Drone Terraform plugin
#
#     docker build --rm=true -t jmccann/drone-terraform:latest .

FROM alpine:3.4

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" | tee -a /etc/apk/repositories && \
  apk -U add \
    ca-certificates \
    git \
    terraform && \
  rm -rf /var/cache/apk/*

ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
