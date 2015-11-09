# Docker image for Drone's terraform deployment plugin
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/drone-terraform .

FROM gliderlabs/alpine:3.2
RUN apk-install ca-certificates
RUN mkdir /terraform
ENV PATH /terraform:$PATH
WORKDIR /terraform
ADD https://releases.hashicorp.com/terraform/0.6.6/terraform_0.6.6_linux_amd64.zip terraform.zip
RUN unzip terraform.zip
WORKDIR /
ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
