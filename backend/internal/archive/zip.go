package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipDir(sourceDir, zipPath string) error {
	info, err := os.Stat(sourceDir)
	if err != nil {
		return fmt.Errorf("源目录不存在: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("源路径不是目录")
	}

	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("创建压缩包失败: %w", err)
	}
	defer out.Close()

	w := zip.NewWriter(out)
	defer w.Close()

	return filepath.Walk(sourceDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		header.Name = rel
		header.Method = zip.Deflate

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		_, err = io.Copy(writer, in)
		return err
	})
}

func SanitizeZipBaseName(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer(
		"\\", "_", "/", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}
