FROM golang:1.13 AS builder

WORKDIR /tmp/randomizer
ENV CGO_ENABLED=0

COPY go.mod go.sum /tmp/randomizer/
COPY vendor/ /tmp/randomizer/vendor/
COPY cmd/ /tmp/randomizer/cmd/
COPY internal/ /tmp/randomizer/internal/
RUN go install -mod=vendor -ldflags="-s -w" -v ./cmd/randomizer-server


FROM alpine:latest

EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/randomizer-server /usr/local/bin/randomizer-server
