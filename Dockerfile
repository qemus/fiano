# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.7-alpine AS builder

ARG MODULE_COMMIT
ARG TARGETOS TARGETARCH

RUN set -eu && \
    apk --no-cache add git && \
    git clone \
      --branch module \
      --single-branch \
      https://github.com/qemus/fiano.git \
      /src && \
    git -C /src checkout --detach "$MODULE_COMMIT" && \
    rm -rf /src/.git /var/cache/apk/*

WORKDIR /src

RUN go test ./...

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
