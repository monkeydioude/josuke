#!/bin/sh
set -a

stopContainer() {
    containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")
    if [ -n "$containerID" ]; then
        echo "[INFO] stopping remaining '$BIN_IMAGE_NAME' container '$containerID'"
        docker stop $containerID > /dev/null
    fi
}

# Expecting optional BIN_IMAGE_NAME env var from Makefile
if [ -z $BIN_IMAGE_NAME ]; then
    echo "[WARN] BIN_IMAGE_NAME not set. Using 'josuke' as default"
    BIN_IMAGE_NAME=josuke
fi

# Bulding docker image for functional tests
BIN_IMAGE_NAME=$BIN_IMAGE_NAME-func-test:latest
if [ -z $(docker images -q "$BIN_IMAGE_NAME") ]; then
    echo "[INFO] building $BIN_IMAGE_NAME image"
    docker build --target build -f build/Dockerfile -t "$BIN_IMAGE_NAME" .
fi

# looping over every test scripts in test/functional directory
for ftest in test/functional/test*.sh; do
    # stopping already running container since  each test
    # might require a different josuke config
    containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")
    if [ -n "$containerID" ]; then
        echo "[INFO] stopping already running '$BIN_IMAGE_NAME' container '$containerID'"
        docker stop $containerID > /dev/null
    fi

    # =.= zZZz let's not wake them up. Running tests silently
    echo "[INFO] running '$ftest' test"
    $ftest

    exitCode=$? 
    # test exited with a status = error
    if [ ! $exitCode = 0 ]; then
        echo "[ERR ] '$ftest' returned '$exitCode'"
        stopContainer
        exit 1
    fi

    echo "[INFO] '$ftest' OK"
done

echo "[INFO] functional tests successful \o/"
stopContainer