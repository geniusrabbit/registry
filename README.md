# Registry service

go get -v github.com/geniusrabbit/registry

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
