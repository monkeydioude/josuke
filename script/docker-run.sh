#!/bin/sh

u_ok_jojo() {
    maxIt=4
    sleepDuration=3
    sleep $sleepDuration
    for d in `seq 1 $maxIt`; do 
        containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")
        if [ -n "$containerID" ] && [ $(docker inspect --format="{{.State.Health.Status}}" $containerID) = "healthy" ]; then
            return 0
        else
            printf "[WARN] container did not start yet. Sleeping for "$sleepDuration"s \n"
            sleep $sleepDuration
        fi
    done
    return 1
}

DEFAULT_PORT=8082

# Expecting optional BIN_IMAGE_NAME env var from Makefile
if [ -z $BIN_IMAGE_NAME ]; then
    echo "[WARN] BIN_IMAGE_NAME not set. Using 'josuke' as default"
    BIN_IMAGE_NAME=josuke
fi

# Expecting mandatory CONF_FILE env var from Makefile
if [ -z $CONF_FILE ]; then
    echo "[ERR ] CONF_FILE not provided"
    exit 1
fi

# Checking image already has a running container
CONTAINER_ID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")
if [ -n "$CONTAINER_ID" ]; then
    echo "[ERR ] container with ID '$CONTAINER_ID' already running for image '$BIN_IMAGE_NAME'"
    exit 1
fi

PORT=$(cat "$CONF_FILE" | docker run -i imega/jq '.port')

if [ $PORT  = "null" ]; then
    echo "[WARN] 'port' not found in conf file $CONF_FILE, using $DEFAULT_PORT"
    PORT=$DEFAULT_PORT
fi

RM_FLAG= 
if [ -n $RM ]; then
    echo "[INFO] RM_FLAG is set, container will run with --rm flag" 
    RM_FLAG=--rm
fi

containerID=$(docker run --network="host" $RM_FLAG -d -v $(pwd):/src -e "CONF_FILE=$CONF_FILE" -e "PORT=$PORT" $BIN_IMAGE_NAME)
echo "[INFO] container running with untruncated ID '$containerID'"

# checking container status
u_ok_jojo
if [ ! $? = 0 ]; then
    echo "[ERR ] container did not start properly (not running or unhealthy)"
    exit 1
fi

echo "[INFO] container running on http://localhost:$PORT"
echo '[INFO] live logs with `make logs` (see `make help`)'