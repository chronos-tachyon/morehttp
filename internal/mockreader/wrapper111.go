package mockreader

import (
	"io"
	"io/fs"
)

type Wrapper111 struct {
	Inner *MockReader
}

func (w Wrapper111) Read(p []byte) (int, error) {
	return w.Inner.Read(p)
}

func (w Wrapper111) Close() error {
	return w.Inner.Close()
}

func (w Wrapper111) ReadAt(p []byte, offset int64) (int, error) {
	return w.Inner.ReadAt(p, offset)
}

func (w Wrapper111) Seek(offset int64, whence int) (int64, error) {
	return w.Inner.Seek(offset, whence)
}

func (w Wrapper111) Stat() (fs.FileInfo, error) {
	return w.Inner.Stat()
}

var (
	_ fs.File     = Wrapper111{}
	_ io.Reader   = Wrapper111{}
	_ io.ReaderAt = Wrapper111{}
	_ io.Seeker   = Wrapper111{}
	_ io.Closer   = Wrapper111{}
)
