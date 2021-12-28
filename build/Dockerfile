## build image ##
FROM golang:1.15-alpine AS build
WORKDIR /src
EXPOSE 8082
RUN apk add openrc curl jq openssl
# Make alpine's sh source this shell profile file on opening a shell
ENV ENV="/etc/profile"
RUN echo 'alias l="ls -la"' >> /etc/profile.d/aliases.sh
HEALTHCHECK --interval=3s --timeout=3s --start-period=1s --retries=3 CMD curl --fail http://localhost:$PORT/healthcheck || exit 1
ENTRYPOINT DOCKER=1 go build -o /out/josuke ./bin/josuke && /src/script/run.sh

## test image ##
FROM build AS test
WORKDIR /src
RUN apk add gcc g++ libc-dev
ENTRYPOINT go test -v ./...
