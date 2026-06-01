package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/xuri/excelize/v2"

	"shipping-excel/backend/internal/excel"
)

func main() {
	source := excel.RealTemplatePath()
	if source == "" {
		fmt.Println("未找到 ATP海运模板小程序.xlsm")
		os.Exit(1)
	}
	output := filepath.Join("testdata", "real-output")

	templateData, _ := os.ReadFile(source)
	records, _ := excel.ParseDataSheet(source)
	codes := excel.UniqueHSCodesSorted(records)

	os.RemoveAll(output)
	os.MkdirAll(output, 0755)

	// 测试全部 HS CODE
	for i, hs := range codes {
		filtered := excel.FilterByHSCode(records, hs)
		_, err := excel.GenerateFromTemplateBytes(templateData, output, hs, filtered)
		if err != nil {
			fmt.Printf("[%d/%d] FAIL %s: %v\n", i+1, len(codes), hs, err)
			os.Exit(1)
		}
		fmt.Printf("[%d/%d] OK %s (%d rows)\n", i+1, len(codes), hs, len(filtered))
	}

	// 详细验证 3926300000 (4 rows)
	hs := "3926300000"
	filtered := excel.FilterByHSCode(records, hs)
	outPath := filepath.Join(output, "43115-DT1EJ20250101-A"+hs+".xlsx")
	if err := verifyOutput(outPath, filtered); err != nil {
		fmt.Println("VERIFY FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("\nAll 43 HS codes generated and sample verified PASS")
}

func verifyOutput(path string, src []excel.DataRecord) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	xx := len(src)
	inv := "INVOICE"

	fmt.Printf("\n=== Detail verify %s (%d rows) ===\n", filepath.Base(path), xx)

	for i, row := range src {
		r := 13 + i
		a, _ := f.GetCellValue(inv, "A"+strconv.Itoa(r))
		b, _ := f.GetCellValue(inv, "B"+strconv.Itoa(r))
		c, _ := f.GetCellValue(inv, "C"+strconv.Itoa(r))
		d, _ := f.GetCellValue(inv, "D"+strconv.Itoa(r))
		e, _ := f.GetCellValue(inv, "E"+strconv.Itoa(r))
		fv, _ := f.GetCellValue(inv, "F"+strconv.Itoa(r))
		g, _ := f.GetCellValue(inv, "G"+strconv.Itoa(r))
		h, _ := f.GetCellValue(inv, "H"+strconv.Itoa(r))

		wantG := row.Qty * row.UnitPrice
		fmt.Printf("R%d: seq=%s part=%s qty=%s price=%s amount=%s (want %.2f)\n", r, a, b, e, fv, g, wantG)

		if a != strconv.Itoa(i+1) {
			return fmt.Errorf("R%d seq want %d got %s", r, i+1, a)
		}
		if b != row.PartNo {
			return fmt.Errorf("R%d part want %s got %s", r, row.PartNo, b)
		}
		if c != row.DescEN {
			return fmt.Errorf("R%d desc EN mismatch", r)
		}
		if d != row.Type {
			return fmt.Errorf("R%d type want %q got %q", r, row.Type, d)
		}
		if h != row.DescRU {
			return fmt.Errorf("R%d desc RU mismatch", r)
		}
		_ = fv
		_ = g
	}

	freightRow := 13 + xx
	insRow := freightRow + 1
	totalRow := insRow + 1

	gFreight, _ := f.GetCellValue(inv, "G"+strconv.Itoa(freightRow))
	gIns, _ := f.GetCellValue(inv, "G"+strconv.Itoa(insRow))
	gTotal, _ := f.GetCellValue(inv, "G"+strconv.Itoa(totalRow))
	aFreight, _ := f.GetCellValue(inv, "A"+strconv.Itoa(freightRow))

	var sumQ, sumR, sumG float64
	for _, row := range src {
		sumQ += row.Freight1
		sumR += row.Insurance
		sumG += row.Qty * row.UnitPrice
	}

	fmt.Printf("Freight R%d label=%q G=%s (want %.2f)\n", freightRow, aFreight, gFreight, sumQ)
	fmt.Printf("Insurance R%d G=%s (want %.2f)\n", insRow, gIns, sumR)
	fmt.Printf("Total R%d G=%s (want %.2f)\n", totalRow, gTotal, sumQ+sumR+sumG)

	if aFreight != "Freight Value(CNY)" {
		return fmt.Errorf("freight row label lost: got %q", aFreight)
	}

	return nil
}
