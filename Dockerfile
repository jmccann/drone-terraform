# Docker image for the Drone Terraform plugin
#
#     docker build -t jmccann/drone-terraform:latest .
FROM golang:1.13-alpine AS builder

RUN apk add --no-cache git

RUN mkdir -p /tmp/drone-terraform
WORKDIR /tmp/drone-terraform

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o /go/bin/drone-terraform

FROM alpine:3.9

RUN apk -U add \
    ca-certificates \
    git \
    wget \
    postgresql-client \
    openssh-client && \
    rm -rf /var/cache/apk/*

ARG terraform_version=0.12.18
RUN wget -q https://releases.hashicorp.com/terraform/${terraform_version}/terraform_${terraform_version}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip -d /bin && \
  rm -f terraform.zip

COPY --from=builder /go/bin/drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
