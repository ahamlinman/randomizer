package randomizer // import "go.alexhamlin.co/randomizer/internal/randomizer"

import "fmt"

// ResultType represents the type of successful result returned by the
// randomizer.
type ResultType int

const (
	// Selection indicates that the randomizer made a random selection from input
	// options.
	Selection ResultType = iota
	// ShowedHelp indicates that the randomizer displayed its help output.
	ShowedHelp
	// ListedGroups indicates that a group list was successfully obtained.
	ListedGroups
	// ShowedGroup indicates that the options of a single group were successfully obtained.
	ShowedGroup
	// SavedGroup indicates that a group was successfully saved.
	SavedGroup
	// DeletedGroup indicates that a group was successfully deleted.
	DeletedGroup
)

// Result represents a successful randomizer operation.
type Result struct {
	resultType ResultType
	message    string
}

// Type returns the type of this result.
func (r Result) Type() ResultType {
	return r.resultType
}

// Message returns the user-friendly output associated with this result.
func (r Result) Message() string {
	return r.message
}

// Error represents an error encountered by the randomizer. It includes
// friendly help messages that can be displayed directly to users when errors
// occur, along with an underlying developer-friendly error that may be useful
// for debugging.
type Error struct {
	cause    error
	helpText string
}

func (e Error) Error() string {
	return e.cause.Error()
}

// Cause returns the underlying developer-friendly error that represents this
// usage error.
func (e Error) Cause() error {
	return e.cause
}

// HelpText returns user-friendly help text associated with this error. While
// the underlying error is more suitable for developer use, the help text may
// be displayed directly to a user.
func (e Error) HelpText() string {
	if e.helpText != "" {
		return e.helpText
	}

	return fmt.Sprintf("Whoops, I had a problem… %v.", e.cause)
}
