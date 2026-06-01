package excel

import (
	"github.com/xuri/excelize/v2"
)

func writeTestSource(path string) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "数据"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"NO.", "ID", "CI NO.", "", "", "", "", "", "", "CASE NO.", "PART No", "HS CODE", "DESC EN", "DESC RU", "QTY", "Unit price", "运费1", "保费"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	// AD列 = 第30列
	f.SetCellValue(sheet, "AD1", "TYPE")

	rows := [][]interface{}{
		{1, "ID1", "CI001", "", "", "", "", "", "", "C1", "P-A1", "8471300000", "Laptop", "Ноутбук", 10, 5000, 100, 50},
		{2, "ID2", "CI001", "", "", "", "", "", "", "C2", "P-A2", "8471300000", "Battery", "Аккум", 20, 200, 100, 50},
		{3, "ID3", "CI002", "", "", "", "", "", "", "C3", "P-B1", "8517120000", "Phone", "Тел", 50, 3000, 200, 80},
		{4, "ID4", "CI003", "", "", "", "", "", "", "C4", "P-C1", "3926909090", "Plastic", "Пластик", 100, 10, 50, 20},
	}
	types := []string{"TYPE-A", "TYPE-A", "TYPE-B", "TYPE-C"}
	for ri, row := range rows {
		for ci, val := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			f.SetCellValue(sheet, cell, val)
		}
		f.SetCellValue(sheet, cellName(30, ri+2), types[ri])
	}
	return f.SaveAs(path)
}

func writeTestTemplate(path string) error {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "INVOICE"
	f.SetSheetName("Sheet1", sheet)
	for col := 1; col <= 8; col++ {
		cell, _ := excelize.CoordinatesToCellName(col, 12)
		f.SetCellValue(sheet, cell, "header")
	}
	return f.SaveAs(path)
}

func cellName(col, row int) string {
	c, _ := excelize.CoordinatesToCellName(col, row)
	return c
}
