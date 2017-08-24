all: test build run

dall: test docker

test:
	go test

build:
	go install
	go build -o bin/josuke ./bin

run:
	./bin/josuke

docker:
	docker build -t josuke .
	docker run -p 6060:8082 josuke