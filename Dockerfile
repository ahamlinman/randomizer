FROM golang:1.11 AS builder

WORKDIR /tmp/randomizer
ENV CGO_ENABLED=0

COPY go.mod go.sum /tmp/randomizer/
RUN go mod download

COPY cmd/ /tmp/randomizer/cmd/
COPY pkg/ /tmp/randomizer/pkg/
RUN go install -mod=readonly -ldflags="-s -w" -v ./cmd/randomizer-server


FROM alpine:latest

EXPOSE 7636
ENTRYPOINT ["/usr/local/bin/randomizer-server"]

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/randomizer-server /usr/local/bin/randomizer-server
