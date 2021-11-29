#!/bin/sh
set -e

confFile=testdata/functional/test-2-valid-config.json
payloadFile=testdata/functional/master-push-github-payload.json

BIN_IMAGE_NAME="$BIN_IMAGE_NAME" CONF_FILE="$confFile" ./script/docker-run-ftest.sh
containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")

docker exec $containerID /src/testdata/send-payload "/src/$confFile" github "/src/$payloadFile"
docker logs $containerID
# should expect this file.
# this is specified by `confFile`, "command" field inside "hook"
logFileSize=$(docker exec $containerID stat -c %s /tmp/hook.log)

if [ $logFileSize -le 1 ]; then
    exit 1
fi
exit 0
