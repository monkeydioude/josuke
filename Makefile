.PHONY: start docker docker-linux

BASE_IMAGE_NAME=josuke
LINUX_DOCKERFILE=build/linux_Dockerfile
DOCKERFILE=$(LINUX_DOCKERFILE)
DEFAULT_BIN_IMAGE_NAME=$(BASE_IMAGE_NAME)-linux
BIN_IMAGE_NAME=$(DEFAULT_BIN_IMAGE_NAME)
# WIN_DOCKERFILE=build/linux_Dockerfile

start:
	@./script/run.sh

docker:
	docker build -f $(DOCKERFILE) -t $(BIN_IMAGE_NAME) .
	docker run --network="host" -d -e "CONF_FILE=$(CONF_FILE)" $(BIN_IMAGE_NAME)

docker-linux:
	DOCKERFILE=$(LINUX_DOCKERFILE) BIN_IMAGE_NAME=$(BASE_IMAGE_NAME)-linux $(MAKE) docker