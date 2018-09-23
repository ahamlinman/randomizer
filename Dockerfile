FROM golang:1.11 AS builder

COPY go.mod go.sum /tmp/randomizer/
COPY cmd/ /tmp/randomizer/cmd/
COPY pkg/ /tmp/randomizer/pkg/
WORKDIR /tmp/randomizer

ENV CGO_ENABLED=0
RUN go install -mod=readonly -ldflags="-s -w" -v \
  ./cmd/randomize \
  ./cmd/slack-randomize-server \
  ./cmd/randomizer-dbtools


FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/randomize /usr/local/bin/randomize
COPY --from=builder /go/bin/slack-randomize-server /usr/local/bin/slack-randomize-server
COPY --from=builder /go/bin/randomizer-dbtools /usr/local/bin/randomizer-dbtools

EXPOSE 7636
