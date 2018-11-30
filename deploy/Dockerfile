FROM alpine:latest
EXPOSE 8080

LABEL maintainer="GeniusRabbitCo (Dmitry Ponomarev <demdxx@gmail.com>)"

COPY /observer /

ENV REGISTRY_DNS=http://registry:8500/dc1?refresh_interval=5
ENV CHECK_HTTP=http://{{address}}/healthcheck
ENV DEFAULT_HOST_IP=0.0.0.0
ENV OBSERVER_REGISTER_HOST_IP=t

CMD ["/observer", "--listen", ":8080"]