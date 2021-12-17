package mockreader

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	NameValue    string
	SizeValue    int64
	ModeValue    fs.FileMode
	ModTimeValue time.Time
}

func (fi *FileInfo) Name() string {
	return fi.NameValue
}

func (fi *FileInfo) Size() int64 {
	return fi.SizeValue
}

func (fi *FileInfo) Mode() fs.FileMode {
	return fi.ModeValue
}

func (fi *FileInfo) ModTime() time.Time {
	return fi.ModTimeValue
}

func (fi *FileInfo) IsDir() bool {
	return fi.ModeValue.IsDir()
}

func (fi *FileInfo) Sys() interface{} {
	return nil
}

var _ fs.FileInfo = (*FileInfo)(nil)
