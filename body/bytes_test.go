package body

import (
	"testing"
)

func TestBytesBody(t *testing.T) {
	b := &bytesBody{data: []byte{'a', 'b', 'c', 'd'}}

	RunBodyTests(t, &TestOptions{
		ShortBody: b,
	})
}
