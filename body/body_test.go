package body

import (
	"testing"
)

func TestFromPrettyJSON(t *testing.T) {
	arr := []int{1, 2, 3}
	b := FromPrettyJSON(arr)

	const CRLF = "\r\n"
	expect := `[` + CRLF + `  1,` + CRLF + `  2,` + CRLF + `  3` + CRLF + `]` + CRLF
	actual := string(b.(*bytesBody).data)
	if expect != actual {
		t.Errorf("expected %q, got %q", expect, actual)
	}
}
