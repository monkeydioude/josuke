.PHONY: install start run stop restart shell test go_start go_test

BIN_IMAGE_NAME=josuke

install:
	git config core.hooksPath $(shell pwd)/.githooks
	@# Not optimal (won't have hooks update unless triggering this rule again),
	@# but in case of not having a git version < 2.9
	cp ./.githooks/* ./.git/hooks/

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