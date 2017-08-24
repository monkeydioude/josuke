FROM golang
MAINTAINER Monkeydioude <monkeydioude@gmail.com>
WORKDIR /go/src/github.com/monkeydioude/josuke
ADD . /go/src/github.com/monkeydioude/josuke
RUN go build -o bin/josuke ./bin
ENTRYPOINT /go/src/github.com/monkeydioude/josuke/bin/josuke
EXPOSE 8082
