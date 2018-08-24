#################################################################################################
# Docker image for the Drone Terraform plugin
#


#################################################################################################
#   Build the Go binary
FROM golang:1.10-alpine AS builder
COPY ./*.go ./src/
COPY ./vendor/ ./src/
RUN set -ex \
    && cd ./src \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o /go/bin/drone-terraform


#################################################################################################
#   Build the Terraform plugin
FROM golang:1.10-alpine AS tfbuilder

RUN apk -U add \
    ca-certificates \
    git \
    bash \
    wget \
    rm -rf /var/cache/apk/*

RUN mkdir -p /go/src/github.com/sl1pm4t \
    && cd /go/src/github.com/sl1pm4t \
    && git clone https://github.com/sl1pm4t/terraform-provider-kubernetes.git \
    && cd ./terraform-provider-kubernetes \
    && go get -v \
    && GOOS=linux GOARCH=amd64 go build -v -o /go/bin/terraform-provider-kubernetes \
    && ls -al /go/bin/terraform-provider-kubernetes

#################################################################################################
# Build the actual container  
FROM alpine:3.7

RUN apk -U add \
    ca-certificates \
    git \
    ansible \
    jq \
    bash \
    wget \
    openssh-client && \
    rm -rf /var/cache/apk/*

ENV TERRAFORM_VERSION 0.11.8
ENV AWS_PROFILE drone-testing

RUN wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip -d /bin && \
  rm -f terraform.zip

RUN wget -P /usr/local/bin/ https://amazon-eks.s3-us-west-2.amazonaws.com/1.10.3/2018-07-26/bin/linux/amd64/aws-iam-authenticator && \
    chmod +x /usr/local/bin/aws-iam-authenticator

RUN wget -P /usr/local/bin/ https://amazon-eks.s3-us-west-2.amazonaws.com/1.10.3/2018-07-26/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

RUN mkdir -p /root/.terraform.d/plugins/

COPY --from=tfbuilder /go/bin/terraform-provider-kubernetes /root/.terraform.d/plugins/
COPY --from=builder /go/bin/drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
