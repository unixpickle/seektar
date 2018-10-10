package seektar

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/unixpickle/essentials"
)

func TestTar(t *testing.T) {
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

	tarObj, err := Tar(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	tarFile, err := tarObj.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)

	entries := []struct {
		Name     string
		Contents string
	}{
		{longFilename, ""},
		{longFilename + "/" + longFilename, "this is a test"},
		{
			longFilename + "/" + longFilename + longFilename,
			longFilename + longFilename + longFilename + longFilename,
		},
		{"file1", "testing"},
		{"file2", "toasting123"},
	}

	for i, entry := range entries {
		header, err := tarReader.Next()
		if err != nil {
			t.Fatalf("file %d: error %s", i, err)
		}
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
}
