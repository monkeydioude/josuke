.PHONY: install start build run stop restart shell bb logs offline_logs attach sr test ftest go_start go_test

BIN_IMAGE_NAME=josuke
RUNNING_CONTAINER_ID=$(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

install:
	git config core.hooksPath $(shell pwd)/.githooks
	@# Not optimal (won't have hooks update unless triggering this rule again),
	@# but in case of having git version < 2.9
	cp ./.githooks/* ./.git/hooks/

help:
	@echo 'List of make actions (stared actions require CONF_FILE=/path/to/conf.json parameter):'
	@echo "\t- install: setup development tooling"
	@echo "\t- build: build $(BIN_IMAGE_NAME) Docker image"
	@echo "\t* run: run $(BIN_IMAGE_NAME) Docker container. Image needs to be built"
	@echo "\t* start: build + run"
	@echo "\t- stop: stop $(BIN_IMAGE_NAME) running Docker container"
	@echo "\t* restart: stop + start"
	@echo "\t- shell: run a shell CLI inside $(BIN_IMAGE_NAME) running Docker container"
	@echo "\t- bb: rebuild $(BIN_IMAGE_NAME)'s binary within an already running Docker container"
	@echo "\t- logs: live display of running $(BIN_IMAGE_NAME) Docker container"
	@echo "\t- offline_logs: allows to display stopped $(BIN_IMAGE_NAME) Docker container's logs"
	@echo "\t- attach: attach a local tty to a running $(BIN_IMAGE_NAME) Docker container. Recovering terminal input might need kill -9"
	@echo "\t* sr: stop + run"
	@echo "\t- test: run unit tests inside a container"
	@echo "\t* ftest: run functional tests without a sequence of containers"
	@echo "\t* go_start: build and run $(BIN_IMAGE_NAME) on local machine"
	@echo "\t- go_test: run unit tests on local machine"

start: build run

build:
	docker build --target build -f build/Dockerfile -t $(BIN_IMAGE_NAME) .

run:
	@BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) CONF_FILE=$(CONF_FILE) ./script/docker-run.sh

stop:
	docker stop $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

restart: stop start

shell:
	docker exec -it $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME)) /bin/sh

bb:
	docker exec -it $(shell docker ps -alqf ancestor=$(BIN_IMAGE_NAME)) go build -o /out/josuke /src/bin/josuke

logs:
	docker exec -it $(RUNNING_CONTAINER_ID) tail -f /var/log/josuke

offline_logs:
	@BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) ./script/docker-container-logs.sh

attach:
	docker attach --detach-keys="d" $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

sr: stop run

test:
	@TEST_IMAGE_NAME=$(BIN_IMAGE_NAME)-test:latest ./script/docker-test.sh

ftest:
	@BIN_IMAGE_NAME="$(BIN_IMAGE_NAME)" ./script/functional-test-runner.sh

go_start:
	@CONF_FILE=$(CONF_FILE) ./script/run.sh

go_test:
	go test -v ./...