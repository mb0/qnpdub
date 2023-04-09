package pcm

import (
	"io"
	"os"
)

type Info struct {
	Format
	Path  string
	Count int
}

type File struct {
	Info
	io.ReadSeekCloser
}

func Open(path string, f Format) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	count := int(fi.Size() / int64(f.Bytes))
	return &File{Info{f, path, count}, file}, nil
}
