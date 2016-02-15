# Docker image for Drone's terraform deployment plugin
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/drone-terraform .

FROM gliderlabs/alpine:3.2
RUN apk-install ca-certificates

ENV TERRAFORM_VERSION 0.6.11

RUN apk update && \
    wget -q "https://circle-artifacts.com/gh/andyshinn/alpine-pkg-glibc/6/artifacts/0/home/ubuntu/alpine-pkg-glibc/packages/x86_64/glibc-2.21-r2.apk" && \
    apk add --allow-untrusted glibc-2.21-r2.apk && \
    wget -q -O terraform.zip "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip terraform.zip -d /bin && \
    rm -rf /var/cache/apk/* glibc-2.21-r2.apk terraform.zip

ADD drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
