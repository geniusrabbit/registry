FROM alpine
MAINTAINER GeniusRabbitCo

COPY ./.build/docker/observer /observer

ENV registry_DNS=http://registry:8500/dc1?refresh_interval=5

ENTRYPOINT /observer
