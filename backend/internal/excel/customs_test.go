package excel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueHSCodesSorted(t *testing.T) {
	records := []DataRecord{
		{HSCode: "8471300000"},
		{HSCode: "8517120000"},
		{HSCode: "8471300000"},
		{HSCode: "3926909090"},
	}
	codes := UniqueHSCodesSorted(records)
	if len(codes) != 3 {
		t.Fatalf("expected 3 unique codes, got %d", len(codes))
	}
	if codes[0] != "3926909090" {
		t.Fatalf("expected sorted first code 3926909090, got %s", codes[0])
	}
}

func TestNextHSCode(t *testing.T) {
	codes := []string{"111", "222", "333"}
	if got := NextHSCode(codes, ""); got != "111" {
		t.Fatalf("expected 111, got %s", got)
	}
	if got := NextHSCode(codes, "111"); got != "222" {
		t.Fatalf("expected 222, got %s", got)
	}
	if got := NextHSCode(codes, "333"); got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

func TestSanitizeFileName(t *testing.T) {
	if got := sanitizeFileName("CI/2024*test"); got != "CI_2024_test" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestGetLastHSCodeFromOutputDir(t *testing.T) {
	dir := t.TempDir()
	files := []string{"ABC8471300000.xlsx", "ABC8517120000.xlsx", "readme.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	last := GetLastHSCodeFromOutputDir(dir)
	if last != "8517120000" {
		t.Fatalf("expected 8517120000, got %s", last)
	}
}

func TestFilterByHSCode(t *testing.T) {
	records := []DataRecord{
		{HSCode: "111", PartNo: "A"},
		{HSCode: "222", PartNo: "B"},
		{HSCode: "111", PartNo: "C"},
	}
	filtered := FilterByHSCode(records, "111")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(filtered))
	}
}
