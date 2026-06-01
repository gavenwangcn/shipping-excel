package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func main() {
	path := `F:\work\shipping-excel\ATP海运模板小程序.xlsm`
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	f, _ := excelize.OpenFile(path)
	defer f.Close()

	for _, sheet := range []string{"INVOICE", "PL"} {
		fmt.Printf("\n=== %s ===\n", sheet)
		for r := 1; r <= 25; r++ {
			var parts []string
			for c := 1; c <= 13; c++ {
				col, _ := excelize.ColumnNumberToName(c)
				v, _ := f.GetCellValue(sheet, col+strconv.Itoa(r))
				if v != "" {
					parts = append(parts, fmt.Sprintf("%s=%q", col, trunc(v, 40)))
				}
			}
			if len(parts) > 0 {
				fmt.Printf("R%d: %v\n", r, parts)
			}
		}
	}

	// sample PL data row mapping from 数据 sheet row 2
	fmt.Println("\n=== 数据 row 2 key cols ===")
	data := "数据"
	cols := map[string]int{
		"K(PART)": 11, "L(HS)": 12, "M(DESC)": 13, "O(QTY)": 15,
		"AE(PKGS)": 31, "AF(TOTAL NW)": 32, "V(GW)": 22, "AG(TOTAL GW)": 33,
		"AH(PKG DIM)": 34, "AB(VOLUME)": 28, "AI(PKG MARK)": 35,
		"AJ(MFG)": 36, "AK(TRADE)": 37,
	}
	for name, col := range cols {
		cell, _ := excelize.CoordinatesToCellName(col, 2)
		v, _ := f.GetCellValue(data, cell)
		fmt.Printf("  %s: %q\n", name, v)
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
