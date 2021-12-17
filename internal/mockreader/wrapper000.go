package mockreader

import (
	"io"
)

type Wrapper000 struct {
	Inner *MockReader
}

func (w Wrapper000) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper000) Close() error {
	return w.Inner.Close()
}

var (
	_ io.Reader = Wrapper000{}
	_ io.Closer = Wrapper000{}
)
