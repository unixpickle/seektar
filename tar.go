package seektar

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
)

type TarTypeFlag byte

const (
	NormalFile TarTypeFlag = 0
	Directory  TarTypeFlag = 5
)

// Tar generates a tarball as a Piece.
//
// If prefix is specified, then it is used as a directory
// name for the tarred content. Otherwise, the tarred
// content is stored relative to the root of the archive.
//
// The result is deterministic, provided that the contents
// of the directory do not change.
func Tar(dirPath, prefix string) (Agg, error) {
	var pieces []Piece

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		filename, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		if filename == "." && prefix == "" {
			return nil
		}
		if prefix != "" {
			filename = filepath.Join(prefix, filename)
		}
		header := &tarHeader{
			Filename: filepath.ToSlash(filename),
			FileMode: uint(info.Mode() & os.ModePerm),
			ModTime:  uint64(info.ModTime().Unix()),
		}
		if info.IsDir() {
			header.Type = Directory
		} else {
			header.FileSize = uint64(info.Size())
			header.Type = NormalFile
		}
		// TODO: get owner information.
		pieces = append(pieces, BytePiece(header.Encode()))
		if !info.IsDir() {
			piece, err := NewFilePiece(path)
			if err != nil {
				return err
			}
			pieces = append(pieces, piece)
			if header.FileSize%512 != 0 {
				padSize := 512 - header.FileSize%512
				pieces = append(pieces, BytePiece(make([]byte, padSize)))
			}
		}
		return nil
	})
	if err != nil {
		return nil, essentials.AddCtx("tar directory", err)
	}
	return Agg(pieces), nil
}

type tarHeader struct {
	Filename    string
	FileMode    uint
	OwnerID     uint
	GroupID     uint
	FileSize    uint64
	ModTime     uint64
	Type        TarTypeFlag
	LinkedFile  string
	OwnerName   string
	GroupName   string
	DeviceMajor uint
	DeviceMinor uint
}

func (t *tarHeader) Encode() []byte {
	var res bytes.Buffer
	filenamePrefix, filenameSuffix := splitFilename(t.Filename)
	padNull(&res, filenameSuffix, 100)
	res.WriteString(fmt.Sprintf("%06o \x00", t.FileMode))
	res.WriteString(fmt.Sprintf("%06o \x00", t.OwnerID))
	res.WriteString(fmt.Sprintf("%06o \x00", t.GroupID))
	res.WriteString(fmt.Sprintf("%11o\x00", t.FileSize))
	res.WriteString(fmt.Sprintf("%11o\x00", t.ModTime))
	res.WriteString("        ")
	res.WriteByte(byte(t.Type))
	padNull(&res, []byte(t.LinkedFile), 100)
	res.WriteString(fmt.Sprintf("ustar\x0000"))
	padNull(&res, []byte(t.OwnerName), 32)
	padNull(&res, []byte(t.GroupName), 32)
	res.WriteString(fmt.Sprintf("%06o \x00", t.DeviceMajor))
	res.WriteString(fmt.Sprintf("%06o \x00", t.DeviceMinor))
	padNull(&res, filenamePrefix, 155)
	for res.Len() < 512 {
		res.WriteByte(0)
	}

	resBytes := res.Bytes()
	var sum uint
	for _, b := range resBytes {
		sum += uint(b)
	}
	copy(resBytes[148:], []byte(fmt.Sprintf("%06o\x00 ", sum)))
	return resBytes
}

func padNull(out *bytes.Buffer, data []byte, length int) {
	if len(data) > length {
		out.Write(data[:length])
	} else if len(data) == length {
		out.Write(data)
	} else {
		out.Write(data)
		for i := len(data); i < length; i++ {
			out.WriteByte(0)
		}
	}
}

func splitFilename(filename string) (prefix, suffix []byte) {
	suffix = []byte(filename)
	if len(suffix) > 100 {
		origData := suffix
		for i, ch := range suffix {
			if i > 155 {
				break
			}
			if ch == filepath.Separator {
				prefix = origData[:i]
				suffix = origData[i+1:]
			}
		}
	}
	return
}
