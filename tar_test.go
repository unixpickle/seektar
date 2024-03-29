package seektar

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestTar(t *testing.T) {
	testTar(t, false, false)
}

func TestTarHTTP(t *testing.T) {
	t.Run("UseRoot", func(t *testing.T) {
		testTar(t, true, true)
	})
	t.Run("NoRoot", func(t *testing.T) {
		testTar(t, true, false)
	})
}

func testTar(t *testing.T, useHTTP, httpUseRoot bool) {
	dir, err := ioutil.TempDir("", "agg_test")
	essentials.Must(err)
	defer os.RemoveAll(dir)

	essentials.Must(ioutil.WriteFile(filepath.Join(dir, "file1"), []byte("testing"), 0654))
	essentials.Must(ioutil.WriteFile(filepath.Join(dir, "file2"), []byte("toasting123"), 0644))

	longFilename := ""
	for i := 0; i < 50; i++ {
		longFilename += "a"
	}

	essentials.Must(os.Mkdir(filepath.Join(dir, longFilename), 0700))
	essentials.Must(ioutil.WriteFile(filepath.Join(dir, longFilename, longFilename),
		[]byte("this is a test"), 0664))
	essentials.Must(ioutil.WriteFile(filepath.Join(dir, longFilename, longFilename+longFilename),
		[]byte(longFilename+longFilename+longFilename+longFilename), 0664))

	for _, prefix := range []string{"", "Dir"} {
		t.Run("Prefix"+prefix, func(t *testing.T) {
			var tarObj Agg
			var err error
			if useHTTP {
				if httpUseRoot {
					tarObj, err = TarHTTP(http.Dir("/"), filepath.ToSlash(dir), prefix)
				} else {
					tarObj, err = TarHTTP(http.Dir(dir), "/", prefix)
				}
			} else {
				tarObj, err = Tar(dir, prefix)
			}
			if err != nil {
				t.Fatal(err)
			}

			tarFile, err := tarObj.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer tarFile.Close()

			testRandomSeekConsistency(t, tarFile)
			if _, err := tarFile.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}

			tarReader := tar.NewReader(tarFile)

			entries := []struct {
				Name     string
				Contents string
			}{
				{"", ""},
				{longFilename, ""},
				{longFilename + "/" + longFilename, "this is a test"},
				{
					longFilename + "/" + longFilename + longFilename,
					longFilename + longFilename + longFilename + longFilename,
				},
				{"file1", "testing"},
				{"file2", "toasting123"},
			}
			if prefix == "" {
				entries = entries[1:]
			}

			for i, entry := range entries {
				header, err := tarReader.Next()
				if err != nil {
					t.Fatalf("file %d: error %s", i, err)
				}
				entry.Name = path.Join(prefix, entry.Name)
				if header.Name != entry.Name {
					t.Errorf("expected name %s but got %s", entry.Name, header.Name)
				}
				data, err := ioutil.ReadAll(tarReader)
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != entry.Contents {
					t.Errorf("expected contents %s but got %s", entry.Contents, string(data))
				}
			}
			if _, err := tarReader.Next(); err != io.EOF {
				t.Error("expected EOF")
			}
		})
	}
}

func testRandomSeekConsistency(t *testing.T, r io.ReadSeeker) {
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, size)
	_, err = io.ReadFull(r, data)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		start := rand.Intn(int(size))
		length := rand.Intn(1 + int(size) - start)
		_, err = r.Seek(int64(start), io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		chunk := make([]byte, length)
		if _, err := io.ReadFull(r, chunk); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(chunk, data[start:start+length]) {
			t.Fatalf("expected chunk %v but got %v", chunk, data[start:start+length])
		}
	}
}
