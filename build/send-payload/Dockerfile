## send-payload image ##
FROM alpine
WORKDIR /src
RUN apk add curl openssl jq
ENTRYPOINT /src/send-payload $0 $1 $2