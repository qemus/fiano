# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.7-alpine AS builder

ARG VERSION_ARG="0.0"
ARG TARGETOS TARGETARCH

COPY lzma-level.patch preserve-unchanged.patch /tmp/

RUN set -eu && \
    apk --no-cache add \
    git && \
    git clone -b "v$VERSION_ARG" https://github.com/linuxboot/fiano.git /src && \
    git -C /src apply --check /tmp/lzma-level.patch && \
    git -C /src apply /tmp/lzma-level.patch && \
    git -C /src apply --recount --check /tmp/preserve-unchanged.patch && \
    git -C /src apply --recount /tmp/preserve-unchanged.patch && \
    rm -rf /tmp/* /var/cache/apk/*

WORKDIR /src/cmds/utk

RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
      -a \
      -installsuffix cgo \
      -o /src/utk \
      .

FROM scratch

COPY --chmod=755 --from=builder /src/utk /utk.bin

ENTRYPOINT ["/utk.bin"]
