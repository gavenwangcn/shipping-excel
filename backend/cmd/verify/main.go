package main

import (
	"fmt"
	"os"
	"path/filepath"

	"shipping-excel/backend/internal/excel"
)

func main() {
	source := filepath.Join("testdata", "source.xlsx")
	template := filepath.Join("testdata", "template.xlsx")
	output := filepath.Join("testdata", "output")

	records, err := excel.ParseDataSheet(source)
	if err != nil {
		fmt.Println("parse error:", err)
		os.Exit(1)
	}
	fmt.Printf("parsed %d records\n", len(records))

	templateData, _ := os.ReadFile(template)
	codes := excel.UniqueHSCodesSorted(records)
	fmt.Printf("HS codes: %v\n", codes)

	os.RemoveAll(output)
	os.MkdirAll(output, 0755)

	for _, hs := range codes {
		filtered := excel.FilterByHSCode(records, hs)
		result, err := excel.GenerateFromTemplateBytes(templateData, output, hs, filtered)
		if err != nil {
			fmt.Printf("generate %s error: %v\n", hs, err)
			os.Exit(1)
		}
		fmt.Printf("generated: %s (%d rows)\n", result.FileName, result.RowCount)
	}
	fmt.Println("done")
}
