package mockreader

import (
	"io"
	"io/fs"
)

type Wrapper110 struct {
	Inner *MockReader
}

func (w Wrapper110) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper110) Close() error {
	return w.Inner.Close()
}

func (w Wrapper110) Seek(offset int64, whence int) (int64, error) {
	return w.Inner.Seek(offset, whence)
}

func (w Wrapper110) Stat() (fs.FileInfo, error) {
	return w.Inner.Stat()
}

var (
	_ fs.File     = Wrapper110{}
	_ io.Reader   = Wrapper110{}
	_ io.Seeker   = Wrapper110{}
	_ io.Closer   = Wrapper110{}
)
