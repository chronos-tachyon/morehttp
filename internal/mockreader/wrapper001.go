package mockreader

import (
	"io"
)

type Wrapper001 struct {
	Inner *MockReader
}

func (w Wrapper001) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper001) Close() error {
	return w.Inner.Close()
}

func (w Wrapper001) ReadAt(p []byte, offset int64) (int, error) {
	return w.Inner.ReadAt(p, offset)
}

var (
	_ io.Reader   = Wrapper001{}
	_ io.ReaderAt = Wrapper001{}
	_ io.Closer   = Wrapper001{}
)
