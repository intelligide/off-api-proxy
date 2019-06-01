FROM golang:1.12 AS builder

WORKDIR /src
COPY . .

ENV BUILD_HOST=github.com
ENV BUILD_USER=docker
RUN rm -f off-proxy && go run build.go -no-upgrade build off-proxy

FROM alpine:3.9

COPY --from=builder /src/build/bin/off-proxy /bin/off-proxy
COPY --from=builder /src/package/docker/config.toml ./config.toml
COPY --from=builder /src/package/docker/docker-entrypoint.sh /bin/entrypoint.sh

ENV PUID=1000 PGID=1000

RUN apk update && apk upgrade && \
    apk add --upgrade ca-certificates su-exec && \
    rm -rf /var/cache/apk/*

EXPOSE 8000
ENTRYPOINT ["/bin/entrypoint.sh", "-p", "8000"]
