package mockreader

import (
	"io"
	"io/fs"
)

type Wrapper101 struct {
	Inner *MockReader
}

func (w Wrapper101) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper101) Close() error {
	return w.Inner.Close()
}

func (w Wrapper101) ReadAt(p []byte, offset int64) (int, error) {
	return w.Inner.ReadAt(p, offset)
}

func (w Wrapper101) Stat() (fs.FileInfo, error) {
	return w.Inner.Stat()
}

var (
	_ fs.File     = Wrapper101{}
	_ io.Reader   = Wrapper101{}
	_ io.ReaderAt = Wrapper101{}
	_ io.Closer   = Wrapper101{}
)
