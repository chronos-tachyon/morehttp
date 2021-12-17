package body

import (
	"testing"
)

func TestEmptyBody(t *testing.T) {
	b := Empty()

	RunBodyTests(t, &TestOptions{
		EmptyBody: b,
	})
}
