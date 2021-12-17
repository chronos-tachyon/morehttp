package mockreader

import (
	"io"
)

type Wrapper010 struct {
	Inner *MockReader
}

func (w Wrapper010) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper010) Close() error {
	return w.Inner.Close()
}

func (w Wrapper010) Seek(offset int64, whence int) (int64, error) {
	return w.Inner.Seek(offset, whence)
}

var (
	_ io.Reader = Wrapper010{}
	_ io.Seeker = Wrapper010{}
	_ io.Closer = Wrapper010{}
)
