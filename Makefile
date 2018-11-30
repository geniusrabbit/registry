
PROJDIR ?= $(CURDIR)/../../../..

build_docker_observer: clean build_observer_with_docker
	cp deploy/Dockerfile .build/Dockerfile
	docker build -t geniusrabbit/service-observer -f .build/Dockerfile .build/

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

clean:
	-rm -fR ./.build

build_observer_with_docker:
	mkdir -p ./.build
	docker run -it --rm --env CGO_ENABLED=0 --env GOPATH="/project" \
		-v="`pwd`/../../../..:/project" -w="/project/src/github.com/geniusrabbit/registry" golang:latest \
		go build -gcflags '-B' -ldflags '-s -w' -o ".build/observer" "observer/docker/main.go"
