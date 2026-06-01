package excel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateZipFromDir(t *testing.T) {
	dir := t.TempDir()
	files := []string{"a.xlsx", "b.xlsx", "readme.txt"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test-"+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := CreateZipFromDir(dir, zipPath); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("zip should not be empty")
	}
}
