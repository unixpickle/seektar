package seektar

import (
	"io"

	"github.com/unixpickle/essentials"
)

// An Agg is a Piece that combines other Pieces.
type Agg []Piece

func (a Agg) Size() int64 {
	var res int64
	for _, p := range a {
		res += p.Size()
	}
	return res
}

func (a Agg) HashID() []byte {
	var id []byte
	for _, p := range a {
		id = append(id, p.HashID()...)
	}
	return id
}

func (a Agg) Open() (ReadSeekCloser, error) {
	return &aggReader{agg: a}, nil
}

type aggReader struct {
	agg    Agg
	offset int64

	reader       ReadSeekCloser
	readerOffset int64
	readerSize   int64
}

func (a *aggReader) Close() error {
	if a.reader != nil {
		return a.reader.Close()
	} else {
		return nil
	}
}

func (a *aggReader) Read(b []byte) (int, error) {
	a.checkReader()
	if a.reader == nil {
		if err := a.openReader(); err != nil {
			if err != io.EOF {
				err = essentials.AddCtx("Agg read", err)
			}
			return 0, err
		}
	}
	amount, err := a.reader.Read(b)
	a.offset += int64(amount)
	return amount, err
}

func (a *aggReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		offset += a.offset
	case io.SeekEnd:
		offset += a.agg.Size()
	}
	a.offset = offset
	return a.offset, nil
}

func (a *aggReader) checkReader() {
	if a.reader == nil {
		if a.offset < a.readerOffset || a.offset >= a.readerOffset+a.readerSize {
			a.reader.Close()
			a.reader = nil
		}
	}
}

func (a *aggReader) openReader() error {
	if a.reader != nil {
		a.reader.Close()
		a.reader = nil
	}
	var offset int64
	for _, p := range a.agg {
		size := p.Size()
		if a.offset >= offset || a.offset < offset+size {
			var err error
			a.reader, err = p.Open()
			if err == nil && a.offset > offset {
				_, err = a.reader.Seek(a.offset-offset, io.SeekStart)
			}
			if err == nil {
				a.readerOffset = offset
				a.readerSize = size
			}
			return err
		}
		offset += size
	}
	return io.EOF
}
