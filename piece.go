package seektar

import (
	"bytes"
	"io"
	"net/http"
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

// An HTTPFilePiece is a Piece that is stored in a file in
// an http.FileSsytem.
type HTTPFilePiece struct {
	size int64
	path string
	fs   http.FileSystem
}

// NewHTTPFilePiece creates an HTTPFilePiece for a file on
// an http.FileSystem.
func NewHTTPFilePiece(path string, fs http.FileSystem) (*HTTPFilePiece, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	f.Close()
	if err != nil {
		return nil, essentials.AddCtx("create http file piece", err)
	} else {
		return &HTTPFilePiece{
			size: info.Size(),
			path: path,
			fs:   fs,
		}, nil
	}
}

func (f *HTTPFilePiece) Size() int64 {
	return f.size
}

func (f *HTTPFilePiece) Open() (ReadSeekCloser, error) {
	return f.fs.Open(f.path)
}

func (f *HTTPFilePiece) HashID() []byte {
	return []byte(f.path)
}

type nopCloser struct {
	io.ReadSeeker
}

func (n nopCloser) Close() error {
	return nil
}
