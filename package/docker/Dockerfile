FROM golang:1.12 AS builder

WORKDIR /src
COPY ../.. .

ENV BUILD_HOST=syncthing.net
ENV BUILD_USER=docker
RUN rm -f off-proxy && go run build.go -no-upgrade build off-proxy

FROM alpine:3.9

COPY --from=builder /src/syncthing /bin/syncthing
COPY --from=builder config.toml ./config.toml
COPY --from=builder /src/script/docker-entrypoint.sh /bin/entrypoint.sh

ENV PUID=1000 PGID=1000

RUN apk update && apt-get upgrade && \
    apk add --upgrade ca-certificates su-exec && \
    rm -rf /var/cache/apk/*


EXPOSE 3000
ENTRYPOINT ["/bin/entrypoint.sh"]
