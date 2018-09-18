FROM golang:1.11 AS builder

COPY . /tmp/randomizer
WORKDIR /tmp/randomizer

ENV CGO_ENABLED=0
RUN go install -mod=readonly -ldflags="-s -w" -v ./cmd/...


FROM busybox:1.29

COPY --from=builder /go/bin/randomize /usr/local/bin/randomize
COPY --from=builder /go/bin/slack-randomize-server /usr/local/bin/slack-randomize-server
COPY --from=builder /go/bin/dynamodb-provision /usr/local/bin/dynamodb-provision

EXPOSE 7636
