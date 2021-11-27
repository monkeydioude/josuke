.PHONY: install start build run stop restart shell bb test logs offline_logs attach sa ra rsa sr go_start go_test

BIN_IMAGE_NAME=josuke
RUNNING_CONTAINER_ID=$(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

install:
	git config core.hooksPath $(shell pwd)/.githooks
	@# Not optimal (won't have hooks update unless triggering this rule again),
	@# but in case of having git version < 2.9
	cp ./.githooks/* ./.git/hooks/

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

test:
	@TEST_IMAGE_NAME=$(BIN_IMAGE_NAME)-test:latest ./script/docker-test.sh

bb:
	docker exec -it $(shell docker ps -alqf ancestor=$(BIN_IMAGE_NAME)) go build -o /out/josuke /src/bin/josuke

logs:
	docker exec -it $(RUNNING_CONTAINER_ID) tail -f /var/log/josuke

offline_logs:
	@BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) ./script/docker-container-logs.sh

attach:
	docker attach --detach-keys="d" $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

sr: stop run

ftest:
	@BIN_IMAGE_NAME="$(BIN_IMAGE_NAME)" ./script/functional-test-runner.sh

go_start:
	@./script/run.sh

go_test:
	go test -v ./...