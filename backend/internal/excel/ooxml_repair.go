package excel

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
)

var calcChainRelPattern = regexp.MustCompile(`<Relationship[^>]*Target="calcChain\.xml"[^>]*/>`)

// repairOOXMLPackage 修复 excelize 写回后可能遗留的无效 calcChain 引用，避免 Excel 打不开。
func repairOOXMLPackage(data []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	hasCalcChain := false
	for _, f := range r.File {
		if f.Name == "xl/calcChain.xml" {
			hasCalcChain = true
			break
		}
	}
	if hasCalcChain {
		return data, nil
	}

	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		if f.Name == "xl/_rels/workbook.xml.rels" {
			content = calcChainRelPattern.ReplaceAll(content, []byte(""))
		}

		hdr := &zip.FileHeader{
			Name:   f.Name,
			Method: f.Method,
		}
		hdr.SetModTime(f.Modified)
		writer, err := w.CreateHeader(hdr)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(content); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
