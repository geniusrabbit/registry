
PROJDIR ?= $(CURDIR)/../../../..

build_docker_observer:
	-rm -f ./.build/docker
	mkdir -p ./.build/docker
	go build -o ./.build/docker/observer observer/docker/main.go
	docker build -t service-observer .

buildenv:
	docker build -t registry_dev .

runenv:
	docker run --rm -it \
		-v $(PROJDIR):/project \
		-v /var/run/docker.sock:/var/run/docker.sock \
		--link consul:registry \
  	registry_dev

run_observer:
	go run observer/docker/main.go

build_observer:
	CGO_ENABLED=0 go build -gcflags '-B' -ldflags '-s -w' -o .build/observer observer/docker/main.go
