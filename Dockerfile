# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG VERSION_ARG="0.0"
ARG TARGETOS TARGETARCH

RUN set -eu && \
    apk --no-cache add \
    git && \
    git clone https://github.com/linuxboot/fiano.git $VERSION_ARG && \
    rm -rf /tmp/* /var/cache/apk/*
    
WORKDIR /$VERSION_ARG

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -o /$VERSION_ARG/utk .

FROM scratch

COPY --chmod=755 --from=builder /$VERSION_ARG/utk /utk.bin

ENTRYPOINT ["/utk.bin"]
