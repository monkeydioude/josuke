#!/bin/sh

# This script build and run josuke binary

BASE_PATH=$(pwd)/

if [ ! -z $DOCKER ]; then
    BASE_PATH=/src/
fi

if [ -z $CONF_FILE ]; then
    echo [ERR ] "A path to a config file must be provided (CONF_FILE env var)"
    exit 1
fi

cd $BASE_PATH"bin/josuke" && go build -o $GOPATH/josuke

$GOPATH/josuke -c $BASE_PATH$CONF_FILE 2> /var/log/josuke