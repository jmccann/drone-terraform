# Docker image for the Drone Terraform plugin
#
#     docker build --rm=true -t jmccann/drone-terraform:latest .
FROM alpine:3.6 as alpine
RUN apk add -U --no-cache ca-certificates git wget

ENV TERRAFORM_VERSION 0.10.3
RUN wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip -O terraform.zip && \
  unzip terraform.zip -d /bin && \
  rm -f terraform.zip

FROM scratch

ENV GODEBUG=netdns=go
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /bin/terraform /bin/terraform



LABEL org.label-schema.version=latest
LABEL org.label-schema.vcs-url="https://github.com/jmccann/drone-terraform.git"
LABEL org.label-schema.name="Drone Terraform"
LABEL org.label-schema.vendor="jmccann"

ADD release/linux/amd64/drone-terraform /bin/
ENTRYPOINT ["/bin/drone-terraform"]
