#!/bin/sh
BIN_IMAGE_NAME="$funcTestImageName" CONF_FILE=./testdata/localhost_test.config.json ./script/docker-run.sh
containerID=$(docker ps -qf ancestor="$funcTestImageName")

docker exec $containerID ls -la /src && exit 0 || exit 1
