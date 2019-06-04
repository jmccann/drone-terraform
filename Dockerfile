
# Docker image for the Drone Terraform plugin
#
#     docker build -t getterminus/drone-terraform:latest .
FROM golang:1.11-alpine AS builder

RUN apk add --no-cache git

RUN mkdir -p /tmp/drone-terraform
WORKDIR /tmp/drone-terraform
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o /go/bin/drone-terraform

FROM alpine:3.9

ENV AWSCLI_VERSION "1.16.52"

RUN apk add --update --no-cache \
    bash \
    python \
    python-dev \
    py-pip \
    build-base \
    && pip install awscli==$AWSCLI_VERSION --upgrade --user \
    && apk --purge -v del py-pip

RUN apk -U --no-cache add \
  ca-certificates \
  curl \
  git \
  openssh-client

ENV INSTALL_DIR /usr/local/bin

ENV TERRAFORM_VERSION 0.11.13
RUN curl -LO https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d ${INSTALL_DIR} && \
  rm -f terraform_${TERRAFORM_VERSION}_linux_amd64.zip

ENV AWS_IAM_AUTHENTICATOR_VERSION 1.12.7/2019-03-27
RUN curl -L -o ${INSTALL_DIR}/aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/${AWS_IAM_AUTHENTICATOR_VERSION}/bin/linux/amd64/aws-iam-authenticator && \
  chmod +x ${INSTALL_DIR}/aws-iam-authenticator

ENV KUBECTL_VERSION v1.14.0
RUN curl -L -o ${INSTALL_DIR}/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
  chmod +x ${INSTALL_DIR}/kubectl

COPY --from=builder /go/bin/drone-terraform ${INSTALL_DIR}/
ENTRYPOINT ["drone-terraform"]

