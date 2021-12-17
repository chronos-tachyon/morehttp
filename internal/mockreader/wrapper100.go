package mockreader

import (
	"io"
	"io/fs"
)

type Wrapper100 struct {
	Inner *MockReader
}

func (w Wrapper100) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper100) Close() error {
	return w.Inner.Close()
}

func (w Wrapper100) Stat() (fs.FileInfo, error) {
	return w.Inner.Stat()
}

var (
	_ fs.File   = Wrapper100{}
	_ io.Reader = Wrapper100{}
	_ io.Closer = Wrapper100{}
)
