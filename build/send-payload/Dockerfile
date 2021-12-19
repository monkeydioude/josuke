## send-payload image ##
FROM alpine
COPY ./testdata/send-payload /send-payload
WORKDIR /src
RUN apk add curl openssl jq
ENTRYPOINT /send-payload $@

# this image should be run like as such:
# docker run --rm -i --network="host" -v $(pwd)/testdata/functional:/src send-payload "/src/config.json" github "/src/payload.json"