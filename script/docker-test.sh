#!/bin/sh

# Expecting optional TEST_IMAGE_NAME env var from Makefile
if [ -z $TEST_IMAGE_NAME ]; then
    echo "[WARN] TEST_IMAGE_NAME not set. Using 'josuke-test:latest' as default"
    TEST_IMAGE_NAME=josuke-test:latest
fi

if [ -z $(docker images -q "$TEST_IMAGE_NAME") ]; then
    docker build --target test -f build/Dockerfile -t "$TEST_IMAGE_NAME" .
fi

docker run -v $(pwd):/src "$TEST_IMAGE_NAME"
