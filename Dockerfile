FROM golang:1.10

COPY . /go/src/go.alexhamlin.co/randomizer

ENV CGO_ENABLED=0
RUN go get -v go.alexhamlin.co/randomizer/cmd/...


FROM alpine:3.8

RUN apk add --no-cache ca-certificates

COPY --from=0 /go/bin/randomize /usr/bin/randomize
COPY --from=0 /go/bin/slack-randomize-server /usr/bin/slack-randomize-server
EXPOSE 7636
