package tmpl

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type templateFile struct {
	name     string
	contents string
}

func createTestDir(t *testing.T, files []templateFile) string {
	dir, err := ioutil.TempDir("", "template")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		fdir := filepath.Dir(file.name)
		if err := os.MkdirAll(filepath.Join(dir, fdir), 0777); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(dir, file.name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, err = io.WriteString(f, file.contents)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFilesTemplate(t *testing.T) {
	dir := createTestDir(t, []templateFile{
		{"base.tmpl", "from a file!"},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}

	tmp := Parse(j("base.tmpl"))
	var buf bytes.Buffer
	if err := tmp.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "from a file!" {
		t.Fatal("\nExp %q\nGot %q", "from a file!", buf.String())
	}
}
