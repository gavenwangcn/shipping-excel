package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

func main() {
	path := `F:\work\shipping-excel\ATP海运模板小程序.xlsm`
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println("open error:", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Println("=== Sheets ===")
	for _, s := range f.GetSheetList() {
		fmt.Println(" -", s)
	}

	// 数据 sheet headers
	dataSheet := "数据"
	rows, _ := f.GetRows(dataSheet)
	fmt.Printf("\n=== %s: %d rows ===\n", dataSheet, len(rows))
	if len(rows) > 0 {
		fmt.Println("Row 1 headers:")
		for i, h := range rows[0] {
			col, _ := excelize.ColumnNumberToName(i + 1)
			if strings.TrimSpace(h) != "" {
				fmt.Printf("  %s(%d): %s\n", col, i, h)
			}
		}
	}

	// HS codes stats
	hsMap := map[string]int{}
	ciMap := map[string]string{}
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		hs := cell(row, 11)
		if hs == "" {
			continue
		}
		hsMap[hs]++
		ciMap[hs] = cell(row, 2)
	}
	fmt.Printf("\n=== HS CODE groups: %d ===\n", len(hsMap))
	for hs, cnt := range hsMap {
		fmt.Printf("  HS=%s CI=%s rows=%d\n", hs, ciMap[hs], cnt)
	}

	// Sample first data row
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if cell(row, 11) != "" {
			fmt.Printf("\n=== Sample data row %d ===\n", i+1)
			cols := []int{2, 10, 11, 12, 13, 14, 15, 16, 17, 29}
			names := []string{"C(CI NO)", "K(PART)", "L(HS)", "M(DESC EN)", "N(DESC RU)", "O(QTY)", "P(PRICE)", "Q(运费)", "R(保费)", "AD(TYPE)"}
			for j, c := range cols {
				col, _ := excelize.ColumnNumberToName(c + 1)
				fmt.Printf("  %s %s: %s\n", col, names[j], cell(row, c))
			}
			break
		}
	}

	// INVOICE sheet structure
	inv := "INVOICE"
	fmt.Printf("\n=== INVOICE rows 10-20 ===\n")
	for r := 10; r <= 25; r++ {
		var parts []string
		for c := 1; c <= 8; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			v, _ := f.GetCellValue(inv, col+fmt.Sprint(r))
			if v != "" {
				parts = append(parts, fmt.Sprintf("%s=%q", col, v))
			}
		}
		if len(parts) > 0 {
			fmt.Printf("  R%d: %s\n", r, strings.Join(parts, " "))
		}
	}

	// Check summary area after row 13 for template defaults
	fmt.Printf("\n=== INVOICE G column rows 13-30 ===\n")
	for r := 13; r <= 30; r++ {
		v, _ := f.GetCellValue(inv, "G"+fmt.Sprint(r))
		if v != "" {
			a, _ := f.GetCellValue(inv, "A"+fmt.Sprint(r))
			fmt.Printf("  R%d A=%q G=%q\n", r, a, v)
		}
	}
}

func cell(row []string, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
