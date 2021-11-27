#!/bin/sh

confFile=testdata/functional/test-1-valid-config.json
payloadFile=testdata/functional/test-1-github-payload.json

# launching container with this specific config
BIN_IMAGE_NAME="$funcTestImageName" CONF_FILE=$confFile RM=1 ./script/docker-run.sh
containerID=$(docker ps -qf ancestor="$funcTestImageName")
if [ $containerID = "" ]; then 
  echo "[ERR ] container not running"
  exit 1
fi
docker logs $containerID
docker exec $containerID /src/testdata/send-payload "/src/$confFile" github "/src/$payloadFile"
docker exec $containerID ls -la /tmp/salut && exit 0 || exit 1
