# Docker image for the Drone Terraform plugin
#
#     docker build --rm=true -t jmccann/drone-terraform:latest .
FROM golang:1.8-alpine AS builder
COPY . .
RUN set -ex \
    && apk add --no-cache git \
    && go-wrapper download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo

FROM alpine:3.4

RUN apk -U add \
    ca-certificates \
    git \
    wget && \
  rm -rf /var/cache/apk/*

ENV TERRAFORM_VERSION 0.9.4
RUN wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip -d /bin && \
  rm -f terraform.zip

COPY --from=/go/src/app/drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
