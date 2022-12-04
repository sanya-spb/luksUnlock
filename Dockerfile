#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go mod download -x

ARG PROJECT=luksUnlock
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG RELEASE=
ARG COMMIT=
ARG BUILD_TIME=
ARG COPYRIGHT="sanya-spb"
ARG CGO_ENABLED=0

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=${CGO_ENABLED} go build \
    -ldflags "-s -w \
    -X ${PROJECT}/pkg/version.version=${RELEASE} \
    -X ${PROJECT}/pkg/version.commit=${COMMIT} \
    -X ${PROJECT}/pkg/version.buildTime=${BUILD_TIME} \
    -X ${PROJECT}/pkg/version.copyright=${COPYRIGHT}" \
    -o /go/bin/app/luksUnlock ./cmd/

#final stage olasw-olasw-openway
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app/luksUnlock /app/
COPY --from=builder /go/src/app/config /app/
RUN adduser -SDH goapp
USER goapp
WORKDIR /app
ENTRYPOINT /app/luksUnlock -config /app/config/config.toml
LABEL Name=luksUnlock
VOLUME ["/app/config"]
