
build_docker_observer:
	-rm -f ./.build/docker
	mkdir -p ./.build/docker
	go build -o ./.build/docker/observer observer/docker/main.go
	docker build -t service-observer .
