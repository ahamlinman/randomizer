package slack

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

type response struct {
	Type responseType `json:"response_type"`
	Text string       `json:"text"`
}

func (r response) WriteTo(w io.Writer) (int64, error) {
	if r == (response{}) {
		return 0, nil
	}

	cw := &countWriter{w, 0}
	err := json.NewEncoder(cw).Encode(&r)
	return cw.n, err
}

type countWriter struct {
	io.Writer
	n int64
}

func (cw *countWriter) Write(p []byte) (int, error) {
	n, err := cw.Writer.Write(p)
	cw.n += int64(n)
	return n, err
}

type responseType int

const (
	typeEphemeral responseType = iota
	typeInChannel
)

func (t responseType) MarshalText() ([]byte, error) {
	switch t {
	case typeEphemeral:
		return []byte("ephemeral"), nil

	case typeInChannel:
		return []byte("in_channel"), nil
	}

	return nil, errors.Errorf("unknown response type code %v", t)
}
