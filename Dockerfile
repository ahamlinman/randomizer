FROM --platform=$BUILDPLATFORM golang:1.15-alpine3.12 AS builder
WORKDIR /tmp/randomizer

ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN go build -v \
  -mod=vendor \
  -trimpath -ldflags="-s -w" \
  ./cmd/randomizer-server


FROM alpine:3.12

EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]

RUN apk add --no-cache ca-certificates

COPY --from=builder /tmp/randomizer/randomizer-server /usr/local/bin/randomizer-server
