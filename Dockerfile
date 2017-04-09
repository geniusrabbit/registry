FROM ubuntu:trusty
MAINTAINER GeniusRabbitCo

COPY ./.build/docker/observer /

ENV REGISTRY_DNS=http://registry:8500/dc1?refresh_interval=5

CMD /observer
