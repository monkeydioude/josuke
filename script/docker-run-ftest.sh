#!/bin/sh

# launching container with this specific config
BIN_IMAGE_NAME="$BIN_IMAGE_NAME" CONF_FILE="$CONF_FILE" RM=1 ./script/docker-run.sh
containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")
if [ $containerID = "" ]; then 
  echo "[ERR ] container not running"
  exit 1
fi