#!/bin/sh

# Expecting optional BIN_IMAGE_NAME env var from Makefile
if [ -z $BIN_IMAGE_NAME ]; then
    echo "[WARN] BIN_IMAGE_NAME not set. Using 'josuke' as default"
    BIN_IMAGE_NAME=josuke
fi

CONTAINER_ID=$(docker ps -aqf ancestor="$BIN_IMAGE_NAME" --no-trunc)

sudo cat /var/lib/docker/containers/$CONTAINER_ID/$CONTAINER_ID-json.log