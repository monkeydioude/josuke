FROM golang:1.15-alpine
WORKDIR /src
COPY . .
EXPOSE 8082

# Make alpine's sh source this shell profile file on opening a shell
ENV ENV="/etc/profile"

RUN go build -o /out/josuke ./bin/josuke
RUN apk add curl
RUN echo 'alias l="ls -la"' >> /etc/profile.d/aliases.sh
# The following RUN command is sponsored by "The Linux Order" and
# its most proficient adept Pyxide.
RUN apk add jq

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 CMD curl --fail http://localhost:$PORT/healthcheck || exit 1
ENTRYPOINT DOCKER=1 /src/script/run.sh