//go:build !windows

package excel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCustomsExcelIntegration(t *testing.T) {
	// 生成临时测试文件
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.xlsx")
	templatePath := filepath.Join(tmpDir, "template.xlsx")
	outputDir := filepath.Join(tmpDir, "output")

	if err := writeTestSource(sourcePath); err != nil {
		t.Fatal(err)
	}
	if err := writeTestTemplate(templatePath); err != nil {
		t.Fatal(err)
	}

	records, err := ParseDataSheet(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	codes := UniqueHSCodesSorted(records)
	if len(codes) != 3 {
		t.Fatalf("expected 3 HS codes, got %d", len(codes))
	}

	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}

	for _, hs := range codes {
		filtered := FilterByHSCode(records, hs)
		result, err := GenerateFromTemplateBytes(templateData, outputDir, hs, filtered)
		if err != nil {
			t.Fatalf("generate %s: %v", hs, err)
		}
		if _, err := os.Stat(result.FilePath); err != nil {
			t.Fatalf("output file not found: %v", err)
		}
	}

	files, _ := os.ReadDir(outputDir)
	if len(files) != 3 {
		t.Fatalf("expected 3 output files, got %d", len(files))
	}
}
