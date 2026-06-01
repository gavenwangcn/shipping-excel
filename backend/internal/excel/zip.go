package excel

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreateZipFromDir(sourceDir, zipPath string) error {
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		return err
	}

	tmpPath := zipPath + ".tmp"
	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("创建压缩包失败: %w", err)
	}

	zw := zip.NewWriter(out)
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		out.Close()
		os.Remove(tmpPath)
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".xlsx" && ext != ".xlsm" {
			continue
		}
		if err := addFileToZip(zw, sourceDir, name); err != nil {
			zw.Close()
			out.Close()
			os.Remove(tmpPath)
			return err
		}
	}

	if err := zw.Close(); err != nil {
		out.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	_ = os.Remove(zipPath)
	return os.Rename(tmpPath, zipPath)
}

func addFileToZip(zw *zip.Writer, dir, name string) error {
	path := filepath.Join(dir, name)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
