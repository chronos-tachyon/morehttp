package body

import (
	"testing"
)

func TestEmptyBody(t *testing.T) {
	b := &emptyBody{}

	RunBodyTests(t, &TestOptions{
		EmptyBody: b,
	})
}
