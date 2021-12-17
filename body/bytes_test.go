package body

import (
	"testing"
)

func TestBytesBody(t *testing.T) {
	b0 := FromBytes(nil)
	b1 := FromBytes([]byte{'a', 'b', 'c', 'd'})

	RunBodyTests(t, &TestOptions{
		EmptyBody: b0,
		ShortBody: b1,
	})
}
