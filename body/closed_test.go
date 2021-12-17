package body

import (
	"testing"
)

func TestClosedBody(t *testing.T) {
	b := &closedBody{}

	RunBodyTests(t, &TestOptions{
		ClosedBody: b,
	})
}
