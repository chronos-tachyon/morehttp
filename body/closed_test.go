package body

import (
	"testing"
)

func TestClosedBody(t *testing.T) {
	b := AlreadyClosed()

	RunBodyTests(t, &TestOptions{
		ClosedBody: b,
	})
}
