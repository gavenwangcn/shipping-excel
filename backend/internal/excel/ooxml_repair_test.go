package excel

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRepairOOXMLPackageRemovesDanglingCalcChain(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	rels := []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId6" Target="calcChain.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/calcChain"/>` +
		`<Relationship Id="rId1" Target="worksheets/sheet1.xml" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"/>` +
		`</Relationships>`)
	f, err := w.Create("xl/_rels/workbook.xml.rels")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(rels); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	fixed, err := repairOOXMLPackage(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	r, err := zip.NewReader(bytes.NewReader(fixed), int64(len(fixed)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range r.File {
		if file.Name != "xl/_rels/workbook.xml.rels" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), "calcChain.xml") {
			t.Fatalf("calcChain rel should be removed: %s", content)
		}
		return
	}
	t.Fatal("workbook.xml.rels not found")
}
