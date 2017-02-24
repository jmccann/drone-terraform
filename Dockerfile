# Docker image for the Drone Terraform plugin
#
#     docker build --rm=true -t jmccann/drone-terraform:latest .

FROM alpine:3.5

ENV TERRAFORM_VERSION 0.8.7

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" | tee -a /etc/apk/repositories && \
  apk -U add \
    ca-certificates \
    git \
	wget && \
  rm -rf /var/cache/apk/* && \
  wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip && \
  rm terraform.zip && \
  mv terraform /bin

ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
