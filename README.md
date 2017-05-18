# Registry service

[![Go Report Card](https://goreportcard.com/badge/github.com/geniusrabbit/registry)](https://goreportcard.com/report/github.com/geniusrabbit/registry)

go get -v github.com/geniusrabbit/registry

## Service environment

```sh
ENV SERVICE_NAME={name}
ENV SERVICE_WEIGHT=1
ENV CHECK_HTTP=http://{{address}}/v1/check
ENV CHECK_INTERVAL=5s
ENV CHECK_TIMEOUT=2s

ENV TAG_{TAG_NAME}={VALUE} => {TAG_NAME}={VALUE}
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
