.PHONY: init start run stop restart shell test go_start go_test

BIN_IMAGE_NAME=josuke

install:
	git config core.hooksPath .githooks

start:
	docker build --target build -f build/Dockerfile -t $(BIN_IMAGE_NAME) .
	$(MAKE) run BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) CONF_FILE=$(CONF_FILE)

run:
	@BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) CONF_FILE=$(CONF_FILE) ./script/docker-run.sh

stop:
	docker stop $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

restart: stop start

shell:
	docker exec -it $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME)) /bin/sh

test:
	@TEST_IMAGE_NAME=$(BIN_IMAGE_NAME)-test:latest ./script/docker-test.sh

go_start:
	@./script/run.sh

go_test:
	go test -v ./...