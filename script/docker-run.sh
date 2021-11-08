#!/bin/sh

DEFAULT_PORT=8081

# Expecting optional BIN_IMAGE_NAME env var from Makefile
if [ -z $BIN_IMAGE_NAME ]; then
    echo "[WARN] BIN_IMAGE_NAME not set. Using `josuke` as default"
    BIN_IMAGE_NAME=josuke
fi

# Expecting mandatory CONF_FILE env var from Makefile
if [ -z $CONF_FILE ]; then
    echo "[ERR ] CONF_FILE not provided"
    exit 1
fi

# Checking container is not already running
CONTAINER_ID=$(docker ps -qf ancestor=$BIN_IMAGE_NAME)
if [ -n $CONTAINER_ID ] && [ ! $CONTAINER_ID = "" ]; then
    echo "[ERR ] Container with ID $CONTAINER_ID already running for image $BIN_IMAGE_NAME"
    exit 1
fi

PORT=$(jq '.port' "$CONF_FILE")

if [ $PORT  = "" ]; then
    echo "[WARN] `port` not found in conf file $CONF_FILE, using $DEFAULT_PORT"
    PORT=$DEFAULT_PORT
fi

docker run --network="host" -d -e "CONF_FILE=$CONF_FILE" -e "PORT=$PORT" $BIN_IMAGE_NAME
sleep 1

if [ -z $(docker ps -qf ancestor=$BIN_IMAGE_NAME) ]; then
    echo "[ERR ] Container did not start"
    exit 1
fi

