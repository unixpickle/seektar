package seektar

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestAgg(t *testing.T) {
	dir, err := ioutil.TempDir("", "agg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	essentials.Must(ioutil.WriteFile(filepath.Join(dir, "file1"), []byte("testing"), 0600))
	essentials.Must(ioutil.WriteFile(filepath.Join(dir, "file2"), []byte("toasting123"), 0600))
	essentials.Must(ioutil.WriteFile(filepath.Join(dir, "file3"), []byte("this is a test"), 0600))

	pieces := Agg{}
	for _, name := range []string{"file1", "file2", "file3"} {
		piece, err := NewFilePiece(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		pieces = append(pieces, piece)
	}

	if pieces.Size() != 32 {
		t.Errorf("expected size 32 but got %d", pieces.Size())
	}
	reader, err := pieces.Open()
	if err != nil {
		t.Fatal(err)
	}

	if size, err := reader.Seek(0, io.SeekEnd); err != nil {
		t.Fatal(err)
	} else if size != pieces.Size() {
		t.Errorf("got size %d but expected size %d", size, pieces.Size())
	}

	if zero, err := reader.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	} else if zero != 0 {
		t.Errorf("expected 0 but got %d", zero)
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "testingtoasting123this is a test" {
		t.Errorf("unexpected data: %s", string(data))
	}

	if seventeen, err := reader.Seek(-15, io.SeekCurrent); err != nil {
		t.Fatal(err)
	} else if seventeen != 17 {
		t.Errorf("expected 17 but got %d", seventeen)
	}

	data, err = ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "3this is a test" {
		t.Errorf("unexpected data: %s", string(data))
	}
}
