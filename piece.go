package seektar

import (
	"bytes"
	"io"
	"os"

	"github.com/unixpickle/essentials"
)

type ReadSeekCloser interface {
	io.Closer
	io.ReadSeeker
}

// A Piece represents a piece of a tarball.
type Piece interface {
	Size() int64
	HashID() []byte
	Open() (ReadSeekCloser, error)
}

// A BytePiece is a Piece of pre-defined data.
type BytePiece []byte

func (b BytePiece) Size() int64 {
	return int64(len(b))
}

func (b BytePiece) Open() (ReadSeekCloser, error) {
	return nopCloser{bytes.NewReader([]byte(b))}, nil
}

func (b BytePiece) HashID() []byte {
	return b
}

// A FilePiece is a Piece that is stored in a file.
type FilePiece struct {
	size int64
	path string
}

// NewFilePiece creates a FilePiece for a file on disk.
func NewFilePiece(path string) (*FilePiece, error) {
	if info, err := os.Stat(path); err != nil {
		return nil, essentials.AddCtx("create file piece", err)
	} else {
		return &FilePiece{
			size: info.Size(),
			path: path,
		}, nil
	}
}

func (f *FilePiece) Size() int64 {
	return f.size
}

func (f *FilePiece) Open() (ReadSeekCloser, error) {
	return os.Open(f.path)
}

func (f *FilePiece) HashID() []byte {
	return []byte(f.path)
}

type nopCloser struct {
	io.ReadSeeker
}

func (n nopCloser) Close() error {
	return nil
}
