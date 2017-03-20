FROM golang:1.8.0-alpine

ENV TERRAFORM_VERSION 0.8.8

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" | tee -a /etc/apk/repositories && \
  apk -U add \
    ca-certificates \
    git \
	wget && \
  rm -rf /var/cache/apk/* && \
  wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip -d /bin && \
  rm -rf /var/cache/apk/* terraform.zip

ADD . /go/src/github.com/jmccann/drone-terraform

WORKDIR /go/src/github.com/jmccann/drone-terraform

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -a -tags netgo

ENTRYPOINT ["/go/bin/drone-terraform"]
