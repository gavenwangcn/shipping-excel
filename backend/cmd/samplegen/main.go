// 生成测试用 Excel 文件，用于验证报关单生成逻辑
// 用法: go run ./cmd/samplegen -out ./testdata
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func main() {
	outDir := flag.String("out", "./testdata", "输出目录")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建目录失败: %v\n", err)
		os.Exit(1)
	}

	if err := createSource(filepath.Join(*outDir, "source.xlsx")); err != nil {
		fmt.Fprintf(os.Stderr, "创建源文件失败: %v\n", err)
		os.Exit(1)
	}
	if err := createTemplate(filepath.Join(*outDir, "template.xlsx")); err != nil {
		fmt.Fprintf(os.Stderr, "创建模板失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已生成测试文件到 %s\n", *outDir)
}

func createSource(path string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "数据"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"NO.", "ID", "CI NO.", "CONTACT NO.", "Container No", "SEAL NO.", "CONTRCT", "ITEM",
		"Modification no.", "CASE NO.", "PART No/ артикул", "HS CODE",
		"DESCRIPTION OF GOODS/ Описание товара на англ.", "DESCRIPTION OF GOODS RUSSIAN/ Описание товара на русском яз.",
		"QTY", "Unit price EXW(CNY)/ Цена за штуку", "运费1", "保费", "UNIT CIP(CNY)",
		"NW（KG） / вес нетто", "TOTAL NW（KG）/ общий вес нетто", "GW（KG）/вес брутто",
		"TOTAL GW（KG) / общий вес брутто", "TOTAL CIP(CNY)", "length", "width", "height",
		"Volume (M)", "港口", "TYPE", "PKGS", "TOTAL NW(KGS)", "TOTAL GW(KGS)",
		"Package dimension（MM）", "Package mark", "MANUFACTURER", "TRADE MARK",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	rows := [][]interface{}{
		{1, "ID001", "CI2024001", "CT001", "CONT001", "SEAL001", "C001", "ITEM1", "M1", "CASE1", "PART-A001", "8471300000", "Laptop Computer", "Ноутбук", 10, 5000, 100, 50, 5150, 2.5, 25, 3, 30, 51500, 30, 20, 5, 0.003, "Shanghai", "TYPE-A", 1, 25, 30, "300x200x50", "MARK1", "MFG-A", "BRAND-A"},
		{2, "ID002", "CI2024001", "CT001", "CONT001", "SEAL001", "C001", "ITEM2", "M2", "CASE2", "PART-A002", "8471300000", "Laptop Battery", "Аккумулятор", 20, 200, 100, 50, 250, 0.5, 10, 0.8, 16, 5000, 10, 8, 3, 0.00024, "Shanghai", "TYPE-A", 1, 10, 16, "100x80x30", "MARK2", "MFG-A", "BRAND-A"},
		{3, "ID003", "CI2024002", "CT002", "CONT002", "SEAL002", "C002", "ITEM3", "M3", "CASE3", "PART-B001", "8517120000", "Mobile Phone", "Мобильный телефон", 50, 3000, 200, 80, 3280, 0.2, 10, 0.3, 15, 164000, 15, 8, 1, 0.00012, "Shenzhen", "TYPE-B", 2, 10, 15, "150x75x8", "MARK3", "MFG-B", "BRAND-B"},
		{4, "ID004", "CI2024002", "CT002", "CONT002", "SEAL002", "C002", "ITEM4", "M4", "CASE4", "PART-B002", "8517120000", "Phone Case", "Чехол", 100, 50, 200, 80, 130, 0.05, 5, 0.08, 8, 6500, 12, 8, 2, 0.000192, "Shenzhen", "TYPE-B", 2, 5, 8, "120x60x15", "MARK4", "MFG-B", "BRAND-B"},
		{5, "ID005", "CI2024003", "CT003", "CONT003", "SEAL003", "C003", "ITEM5", "M5", "CASE5", "PART-C001", "3926909090", "Plastic Parts", "Пластиковые детали", 200, 10, 50, 20, 60, 0.01, 2, 0.02, 4, 12000, 5, 5, 2, 0.00005, "Ningbo", "TYPE-C", 5, 2, 4, "50x50x20", "MARK5", "MFG-C", "BRAND-C"},
	}

	for ri, row := range rows {
		for ci, val := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	return f.SaveAs(path)
}

func createTemplate(path string) error {
	f := excelize.NewFile()
	defer f.Close()

	invoice := "INVOICE"
	f.SetSheetName("Sheet1", invoice)

	// 模板头部 (1-12行)
	f.SetCellValue(invoice, "A1", "COMMERCIAL INVOICE")
	f.SetCellValue(invoice, "A12", "No.")
	f.SetCellValue(invoice, "B12", "PART No")
	f.SetCellValue(invoice, "C12", "Description")
	f.SetCellValue(invoice, "D12", "Type")
	f.SetCellValue(invoice, "E12", "QTY")
	f.SetCellValue(invoice, "F12", "Unit Price")
	f.SetCellValue(invoice, "G12", "Amount")
	f.SetCellValue(invoice, "H12", "Description RU")

	// 第13行作为数据模板行
	for col := 1; col <= 8; col++ {
		cell, _ := excelize.CoordinatesToCellName(col, 13)
		f.SetCellValue(invoice, cell, "")
	}

	// 添加 PL 和数据 sheet
	f.NewSheet("PL")
	f.NewSheet("数据")

	return f.SaveAs(path)
}
