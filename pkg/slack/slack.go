package slack

import (
	"encoding/json"
	"fmt"
	"io"
)

// ResponseType represents the manner in which a response to a Slack slash
// command request will be displayed to a user.
type ResponseType int

const (
	// TypeEphemeral causes a response to be displayed to the user only. The
	// slash command invocation will be hidden from others.
	TypeEphemeral ResponseType = iota

	// TypeInChannel causes a response to be displayed in the channel to other
	// Slack users, along with the slash command invocation that triggered it.
	TypeInChannel
)

// MarshalText encodes a ResponseType into the textual representation
// understood by the Slack API.
func (t ResponseType) MarshalText() ([]byte, error) {
	switch t {
	case TypeEphemeral:
		return []byte("ephemeral"), nil

	case TypeInChannel:
		return []byte("in_channel"), nil
	}

	panic(fmt.Errorf("unknown response type code %v", t))
}

// Response represents a response to a Slack slash command.
type Response struct {
	Type ResponseType `json:"response_type"`
	Text string       `json:"text"`
}

// Send writes the JSON form of a Response to the provided writer.
func (r Response) Send(w io.Writer) {
	json.NewEncoder(w).Encode(&r)
}
