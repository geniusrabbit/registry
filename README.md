# Registry service

## Build observer service

```sh
make build_docker_observer
```

Run service
```sh
docker run -itd --restart always \
   -v /var/run/docker.sock:/var/run/docker.sock \
   --name=service-observer
```
