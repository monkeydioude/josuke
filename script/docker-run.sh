#!/bin/sh

u_ok_jojo() {
    maxIt=4
    sleepDuration=3
    sleep $sleepDuration
    for d in `seq 1 $maxIt`; do 
        containerID=$(docker ps -qf ancestor="$BIN_IMAGE_NAME")
        if [ ! $containerID = "" ] && [ $(docker inspect --format="{{.State.Health.Status}}" $containerID) = "healthy" ]; then
            return 0
        else
            printf "[WARN] container did not start yet. Sleeping "$sleepDuration"s \n"
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
if [ ! $CONTAINER_ID = "" ]; then
    echo "[ERR ] Container with ID '$CONTAINER_ID' already running for image '$BIN_IMAGE_NAME'"
    exit 1
fi

PORT=$(jq '.port' "$CONF_FILE")

if [ $PORT  = "null" ]; then
    echo "[WARN] 'port' not found in conf file $CONF_FILE, using $DEFAULT_PORT"
    PORT=$DEFAULT_PORT
fi

docker run --log-driver syslog --network="host" -d -e "CONF_FILE=$CONF_FILE" -e "PORT=$PORT" $BIN_IMAGE_NAME

# checking container status
u_ok_jojo
if [ $? = 1 ]; then
    echo "[ERR ] Container did not start properly (not running or unhealthy)"
    exit 1
fi

echo "[INFO] Container running on http://localhost:$PORT"