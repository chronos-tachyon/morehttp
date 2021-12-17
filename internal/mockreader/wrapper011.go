package mockreader

import (
	"io"
)

type Wrapper011 struct {
	Inner *MockReader
}

func (w Wrapper011) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper011) Close() error {
	return w.Inner.Close()
}

func (w Wrapper011) ReadAt(p []byte, offset int64) (int, error) {
	return w.Inner.ReadAt(p, offset)
}

func (w Wrapper011) Seek(offset int64, whence int) (int64, error) {
	return w.Inner.Seek(offset, whence)
}

var (
	_ io.Reader   = Wrapper011{}
	_ io.ReaderAt = Wrapper011{}
	_ io.Seeker   = Wrapper011{}
	_ io.Closer   = Wrapper011{}
)
