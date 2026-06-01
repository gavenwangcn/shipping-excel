package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZipDir(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "a.xlsx"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "b.xlsx"), []byte("world"), 0644)

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := ZipDir(src, zipPath); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("zip file empty")
	}
}
