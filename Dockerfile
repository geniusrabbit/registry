FROM golang:1.8.3
MAINTAINER GeniusRabbitCo

ENV GOPATH=/project
ENV PATH="$PATH:/project/bin"

WORKDIR /project/src/github.com/geniusrabbit/registry
ENV REGISTRY_DNS=http://registry:8500/dc1?refresh_interval=5
