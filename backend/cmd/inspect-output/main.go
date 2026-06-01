package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func main() {
	path := filepath.Join("testdata", "real-output", "43115-DT1EJ20250101-A3926300000.xlsx")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	defer f.Close()

	inv := "INVOICE"
	for r := 12; r <= 20; r++ {
		var parts []string
		for c := 1; c <= 8; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			v, _ := f.GetCellValue(inv, col+strconv.Itoa(r))
			if v != "" {
				parts = append(parts, fmt.Sprintf("%s=%q", col, v))
			}
		}
		fmt.Printf("R%d: %s\n", r, join(parts))
	}
}

func join(s []string) string {
	if len(s) == 0 {
		return "(empty)"
	}
	out := s[0]
	for i := 1; i < len(s); i++ {
		out += " " + s[i]
	}
	return out
}
