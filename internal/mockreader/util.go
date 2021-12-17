package mockreader

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
)

var _ = fmt.Stringer(nil)
var _ = os.Stderr

func debug(format string, args ...interface{}) {
	//fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
}

func whenceString(whence int) string {
	switch whence {
	case io.SeekStart:
		return "io.SeekStart"
	case io.SeekCurrent:
		return "io.SeekCurrent"
	case io.SeekEnd:
		return "io.SeekEnd"
	default:
		return formatInt(whence)
	}
}

func formatInt(x int) string {
	return formatInt64(int64(x))
}

func formatInt64(x int64) string {
	return strconv.FormatInt(x, 10)
}

func formatAny(x interface{}) string {
	if x == nil {
		return "nil"
	}
	return fmt.Sprintf("%T[%+v]", x, x)
}

func formatBytesTo(buf *bytes.Buffer, p []byte) {
	buf.WriteByte('[')
	for i, ch := range p {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, "%#02x", ch)
	}
	buf.WriteByte(']')
}
