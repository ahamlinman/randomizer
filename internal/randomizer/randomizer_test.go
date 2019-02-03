package randomizer

import (
	"testing"

	"github.com/pkg/errors"
)

var errOriginalCause = errors.New("there was a test error")

func TestErrorCause(t *testing.T) {
	err := Error{
		cause: errOriginalCause,
	}

	if err.Cause() != errOriginalCause {
		t.Errorf("got cause %v, want %v", err.Cause(), errOriginalCause)
	}

	if err.Error() != err.Cause().Error() {
		t.Errorf("got error text %q, want %q", err.Error(), err.Cause().Error())
	}

}

func TestErrorWithHelpText(t *testing.T) {
	const helpText = "Something went wrong."

	err := Error{
		cause:    errOriginalCause,
		helpText: helpText,
	}

	if err.HelpText() != helpText {
		t.Errorf("got help text %q, want %q", err.HelpText(), helpText)
	}
}

func TestErrorWithoutHelpText(t *testing.T) {
	err := Error{
		cause: errOriginalCause,
	}

	const expectedHelpText = "Whoops, I had a problem… there was a test error."

	if err.HelpText() != expectedHelpText {
		t.Errorf("got help text %q, want %q", err.HelpText(), expectedHelpText)
	}
}
