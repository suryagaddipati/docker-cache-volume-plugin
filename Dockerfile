FROM golang:1.9-alpine as builder
COPY . /go/src/github.com/suryagaddipati/docker-cache-volume-plugin
WORKDIR /go/src/github.com/suryagaddipati/docker-cache-volume-plugin
RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
    gcc libc-dev \
    && go install --ldflags '-extldflags "-static"' \
    && apk del .build-deps
CMD ["/go/bin/docker-cache-volume-plugin"]

FROM alpine
RUN apk update && apk add sshfs
RUN mkdir -p /run/docker/plugins /mnt/state /mnt/volumes /mnt/cache/merged/
COPY --from=builder /go/bin/docker-cache-volume-plugin .
CMD ["docker-cache-volume-plugin"]

