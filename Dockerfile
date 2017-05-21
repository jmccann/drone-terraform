# Docker image for the Drone Terraform plugin
#
#     docker build --rm=true -t jmccann/drone-terraform:latest .

FROM alpine:3.4

RUN apk -U add \
    ca-certificates \
    git \
    wget && \
  rm -rf /var/cache/apk/*

ENV TERRAFORM_VERSION 0.9.5
RUN wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip -d /bin && \
  rm -f terraform.zip

ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
