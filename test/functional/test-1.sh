#!/bin/sh
set -e

confFile=testdata/functional/test-1-valid-config.json
payloadFile=testdata/functional/master-push-github-payload.json

BIN_IMAGE_NAME="$BIN_IMAGE_NAME" CONF_FILE="$confFile" ./script/docker-run-ftest.sh
containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")

docker exec $containerID /src/testdata/send-payload "/src/$confFile" github "/src/$payloadFile"
# should expect this file.
# this is specified by `confFile`, "commands" field inside "deployment"
docker exec $containerID ls -la /tmp/salut && exit 0 || exit 1
