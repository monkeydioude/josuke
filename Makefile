.PHONY: start docker-start

BIN_IMAGE_NAME=josuke

start:
	@./script/run.sh

docker-start:
	docker build -f build/Dockerfile -t $(BIN_IMAGE_NAME) .
	docker run --network="host" -d -e "CONF_FILE=$(CONF_FILE)" $(BIN_IMAGE_NAME)