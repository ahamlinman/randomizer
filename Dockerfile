FROM --platform=$BUILDPLATFORM golang:1.15-alpine3.12 AS builder
WORKDIR /tmp/randomizer

ARG TARGETPLATFORM
ENV TARGETPLATFORM=$TARGETPLATFORM

COPY build-internal.sh go.mod go.sum ./
COPY vendor/ ./vendor/
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN ./build-internal.sh


FROM alpine:3.12

EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]

RUN apk add --no-cache ca-certificates

COPY --from=builder /tmp/randomizer/randomizer-server /usr/local/bin/randomizer-server
