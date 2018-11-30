# Registry service [![Go Report Card](https://goreportcard.com/badge/github.com/geniusrabbit/registry)](https://goreportcard.com/report/github.com/geniusrabbit/registry)

 > @license Apache-2.0

go get -v github.com/geniusrabbit/registry

## Service observer environment

To automatic register any docker service by observer in consul you have to configure your container.
Recommended to use docker LABELs.

```dockerfile
LABEL maintainer="..."

LABEL service.name={somename}
LABEL service.weight=1
LABEL service.port=8080 # Used as default in address ip:port
LABEL service.public="true"
# Healhcheck options
LABEL service.check.interval=5s
LABEL service.check.timeout=10s
LABEL service.check.httpaddr=http://{{address}}/v1/check
# Tags
LABEL service.tag_{TAG_NAME}={VALUE} # => {TAG_NAME}={VALUE}
```

Also available environment variables.
```dockerfile
ENV SERVICE_NAME={somename}
ENV SERVICE_PORT=8080
ENV SERVICE_WEIGHT=1
# Healhcheck options
ENV CHECK_HTTP=http://{{address}}/v1/check
ENV CHECK_INTERVAL=5s
ENV CHECK_TIMEOUT=2s
# Tags
ENV TAG_{TAG_NAME}={VALUE} # => {TAG_NAME}={VALUE}
```

## Example of your service Dockerfile

```dockerfile
FROM ubuntu:trusty

LABEL service.name=archivarius
LABEL service.weight=1
LABEL service.port=8080

LABEL service.check.interval=5s
LABEL service.check.timeout=2s
# {{address}} automaticaly replaced to real address of service
LABEL service.check.httpaddr=http://{{address}}/v1/check

EXPOSE {port}, ...
```

## Build observer service

```sh
make build_docker_observer
```

Run service
```sh
docker run -itd --restart always \
   --name=service-observer \
   --link consul:registry \
   -v /var/run/docker.sock:/var/run/docker.sock \
   service-observer
```
