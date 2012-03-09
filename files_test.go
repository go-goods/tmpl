package tmpl

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type fataler interface {
	Fatal(...interface{})
}

type templateFile struct {
	name     string
	contents string
}

func createTestDir(t fataler, files []templateFile) string {
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

	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		if err := tmp.Execute(&buf, nil); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "from a file!" {
			t.Fatalf("\nExp %q\nGot %q", "from a file!", got)
		}
	}
}

func TestFilesTemplateWithBlocks(t *testing.T) {
	dir := createTestDir(t, []templateFile{
		{"base.tmpl", `{% evoke foo %}`},
		{"foo.block", `{% block foo %}bar{% end block %}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}

	tmp := Parse(j("base.tmpl"))
	tmp.Blocks(j("*.block"))

	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		if err := tmp.Execute(&buf, nil); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "bar" {
			t.Fatalf("\nExp %q\nGot %q", "bar", got)
		}
	}
}

func TestFilesTemplateWithTempBlocks(t *testing.T) {
	dir := createTestDir(t, []templateFile{
		{"base.tmpl", `{% evoke foo %}`},
		{"foo.block", `{% block foo %}bar{% end block %}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}

	tmp := Parse(j("base.tmpl"))

	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		if err := tmp.Execute(&buf, nil, j("*.block")); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "bar" {
			t.Fatalf("\nExp %q\nGot %q", "bar", got)
		}
	}

	//ensure test block doesn't stay
	if err := tmp.Execute(ioutil.Discard, nil); err == nil {
		t.Fatal("Expected error with no block definition")
	}
}

func TestFilesTemplateWithBothBlocks(t *testing.T) {
	dir := createTestDir(t, []templateFile{
		{"base.tmpl", `{% evoke foo %}{% evoke bar %}`},
		{"foo.block", `{% block foo %}bar{% end block %}`},
		{"bar.block", `{% block bar %}baz{% end block %}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}

	tmp := Parse(j("base.tmpl"))
	tmp.Blocks(j("foo.block"))

	for i := 0; i < 10; i++ {
		var buf bytes.Buffer
		if err := tmp.Execute(&buf, nil, j("bar.block")); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "barbaz" {
			t.Fatalf("\nExp %q\nGot %q", "barbaz", got)
		}
	}

	//ensure test block doesn't stay
	if err := tmp.Execute(ioutil.Discard, nil); err == nil {
		t.Fatal("Expected error with no block definition")
	}
}
