.PHONY: start run stop restart go_start

BIN_IMAGE_NAME=josuke

start:
	docker build -f build/Dockerfile -t $(BIN_IMAGE_NAME) .
	$(MAKE) run BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) CONF_FILE=$(CONF_FILE)

run:
	@BIN_IMAGE_NAME=$(BIN_IMAGE_NAME) CONF_FILE=$(CONF_FILE) ./script/docker-run.sh

stop:
	docker stop $(shell docker ps -qf ancestor=$(BIN_IMAGE_NAME))

restart: stop start

go_start:
	@./script/run.sh