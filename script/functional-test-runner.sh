#!/bin/sh
set -e
set -a

# Expecting optional BIN_IMAGE_NAME env var from Makefile
if [ -z $BIN_IMAGE_NAME ]; then
    echo "[WARN] BIN_IMAGE_NAME not set. Using 'josuke' as default"
    BIN_IMAGE_NAME=josuke
fi

# Bulding docker image for functional tests
funcTestImageName=$BIN_IMAGE_NAME-func-test:latest
if [ -z $(docker images -q "$funcTestImageName") ]; then
    echo "[INFO] building $funcTestImageName image"
    docker build --target build -f build/Dockerfile -t "$funcTestImageName" .
fi

# looping over every test scripts in test/functional directory
for ftest in test/functional/*.sh; do
    # stopping already running container since  each test
    # might require a different josuke config
    containerID=$(docker ps -qf ancestor="$funcTestImageName")
    if [ ! $containerID = "" ]; then
        echo "[INFO] stopping already running '$funcTestImageName' container '$containerID'"
        docker stop $containerID > /dev/null
    fi

    # =.= zZZz let's not wake them up. Running tests silently
    echo "[INFO] running '$ftest' test"
    $ftest

    # test exited with a status = error
    if [ ! $? = 0 ]; then
        echo "[ERR ] '$ftest' returned '$?'"
        exit 1
    fi

    echo "[INFO] '$ftest' OK"
done

echo "[INFO] functional tests successful \o/"

containerID=$(docker ps -qf ancestor="$funcTestImageName")
if [ ! $containerID = "" ]; then
    echo "[INFO] stopping remaining '$funcTestImageName' container '$containerID'"
    docker stop $containerID > /dev/null
fi