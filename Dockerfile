FROM golang:1.15-alpine AS builder

WORKDIR /tmp/randomizer

COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go install -v \
  -mod=vendor \
  -ldflags="-s -w" \
  ./cmd/randomizer-server


FROM alpine:latest

EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/randomizer-server /usr/local/bin/randomizer-server
