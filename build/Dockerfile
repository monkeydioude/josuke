FROM golang:1.15-alpine
WORKDIR /src
COPY . .
RUN go build -o /out/josuke ./bin/josuke
ENTRYPOINT DOCKER=1 /src/script/run.sh