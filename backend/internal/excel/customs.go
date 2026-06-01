package excel

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	SheetInvoice = "INVOICE"
	SheetPL      = "PL"
	SheetData    = "数据"

	DataStartRow     = 2
	InvoiceDataRow   = 13
	InvoiceInsertRow = 14

	PLDataRow   = 13
	PLInsertRow = 14
)

type DataRecord struct {
	RowNum      int
	CINo        string
	PartNo      string
	HSCode      string
	DescEN      string
	DescRU      string
	Qty         float64
	UnitPrice   float64
	Freight1    float64
	Insurance   float64
	Type        string
	Pkgs        float64
	TotalNW     float64
	GW          float64
	TotalGW     float64
	PkgDim      string
	Volume      float64
	PkgMark     string
	Manufacturer string
	TradeMark   string
}

type GenerateResult struct {
	FileName string
	FilePath string
	HSCode   string
	CINo     string
	RowCount int
}

var hsCodeFilePattern = regexp.MustCompile(`(.+?)(\d{4,12})$`)

func ParseDataSheet(filePath string) ([]DataRecord, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开Excel失败: %w", err)
	}
	defer f.Close()

	sheetName := resolveSheetName(f, SheetData)
	if sheetName == "" {
		return nil, fmt.Errorf("未找到数据表(数据)")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取数据表失败: %w", err)
	}

	var records []DataRecord
	for i := DataStartRow; i <= len(rows); i++ {
		hsCode := getCellStr(f, sheetName, i, 12)
		if hsCode == "" {
			continue
		}
		records = append(records, DataRecord{
			RowNum:       i,
			CINo:         getCellStr(f, sheetName, i, 3),
			PartNo:       getCellStr(f, sheetName, i, 11),
			HSCode:       hsCode,
			DescEN:       getCellStr(f, sheetName, i, 13),
			DescRU:       getCellStr(f, sheetName, i, 14),
			Qty:          getCellFloat(f, sheetName, i, 15),
			UnitPrice:    getCellFloat(f, sheetName, i, 16),
			Freight1:     getCellFloat(f, sheetName, i, 17),
			Insurance:    getCellFloat(f, sheetName, i, 18),
			Type:         getCellStr(f, sheetName, i, 30),
			Pkgs:         getCellFloat(f, sheetName, i, 31),
			TotalNW:      getCellFloat(f, sheetName, i, 32),
			GW:           getCellFloat(f, sheetName, i, 22),
			TotalGW:      getCellFloat(f, sheetName, i, 33),
			PkgDim:       getCellStr(f, sheetName, i, 34),
			Volume:       getCellFloat(f, sheetName, i, 28),
			PkgMark:      getCellStr(f, sheetName, i, 35),
			Manufacturer: getCellStr(f, sheetName, i, 36),
			TradeMark:    getCellStr(f, sheetName, i, 37),
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].HSCode < records[j].HSCode
	})
	return records, nil
}

func GenerateFromTemplateBytes(templateData []byte, outputDir, hsCode string, rows []DataRecord) (*GenerateResult, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("HS CODE %s 无数据行", hsCode)
	}

	ciNo := sanitizeFileName(rows[0].CINo)
	fileName := fmt.Sprintf("%s%s.xlsm", ciNo, hsCode)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}
	outPath := filepath.Join(outputDir, fileName)

	f, err := excelize.OpenReader(bytes.NewReader(templateData))
	if err != nil {
		return nil, fmt.Errorf("打开模板失败: %w", err)
	}

	if err := fillInvoiceSheet(f, rows, hsCode); err != nil {
		f.Close()
		return nil, err
	}
	if err := fillPLSheet(f, rows); err != nil {
		f.Close()
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("生成Excel缓冲失败: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("关闭工作簿失败: %w", err)
	}

	raw := buf.Bytes()
	fixed, err := repairOOXMLPackage(raw)
	if err != nil {
		return nil, fmt.Errorf("修复 Excel 包结构失败: %w", err)
	}

	_ = os.Remove(outPath)
	if err := writeFileWithRetry(outPath, fixed); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	return &GenerateResult{
		FileName: fileName,
		FilePath: outPath,
		HSCode:   hsCode,
		CINo:     rows[0].CINo,
		RowCount: len(rows),
	}, nil
}

func fillInvoiceSheet(f *excelize.File, rows []DataRecord, hsCode string) error {
	sheet := resolveSheetName(f, SheetInvoice)
	if sheet == "" {
		return fmt.Errorf("模板中未找到 INVOICE 表")
	}

	xx := len(rows)
	if err := prepareDataRows(f, sheet, InvoiceDataRow, InvoiceInsertRow, xx); err != nil {
		return err
	}

	var sumG float64
	for i, row := range rows {
		r := InvoiceDataRow + i
		amount := round2(row.Qty * row.UnitPrice)
		setCell(f, sheet, cellRef(1, r), i+1)
		setCell(f, sheet, cellRef(2, r), row.PartNo)
		setCell(f, sheet, cellRef(3, r), row.DescEN)
		setCell(f, sheet, cellRef(4, r), row.Type)
		setCell(f, sheet, cellRef(5, r), row.Qty)
		setCell(f, sheet, cellRef(6, r), row.UnitPrice)
		setCell(f, sheet, cellRef(7, r), amount)
		setCell(f, sheet, cellRef(8, r), row.DescRU)
		setCell(f, sheet, cellRef(9, r), hsCode)
		sumG += amount
	}

	freightRow := InvoiceDataRow + xx
	sumFreight := sumColumn(rows, func(r DataRecord) float64 { return r.Freight1 })
	sumInsurance := sumColumn(rows, func(r DataRecord) float64 { return r.Insurance })

	setCell(f, sheet, cellRef(7, freightRow), round2(sumFreight))
	setCell(f, sheet, cellRef(7, freightRow+1), round2(sumInsurance))
	setCell(f, sheet, cellRef(7, freightRow+2), round2(sumFreight+sumInsurance+sumG))
	return nil
}

func fillPLSheet(f *excelize.File, rows []DataRecord) error {
	sheet := resolveSheetName(f, SheetPL)
	if sheet == "" {
		return fmt.Errorf("模板中未找到 PL 表")
	}

	xx := len(rows)
	totalRow := findTotalRow(f, sheet, PLDataRow+1)
	if totalRow == 0 {
		return fmt.Errorf("PL 表未找到 TOTAL 汇总行")
	}

	templateDataRows := totalRow - PLDataRow
	for i := 0; i < templateDataRows-1; i++ {
		if err := f.RemoveRow(sheet, PLDataRow+1); err != nil {
			return fmt.Errorf("清理 PL 模板样例行失败: %w", err)
		}
	}

	insertCount := xx - 1
	if insertCount > 0 {
		if err := f.InsertRows(sheet, PLInsertRow, insertCount); err != nil {
			return fmt.Errorf("PL 表插入行失败: %w", err)
		}
		copyRowStyle(f, sheet, PLDataRow, PLInsertRow, insertCount, 13)
	}

	gwValues := calcPLGWValues(rows)
	for i, row := range rows {
		r := PLDataRow + i
		setCell(f, sheet, cellRef(1, r), i+1)
		setCell(f, sheet, cellRef(2, r), row.PartNo)
		setCell(f, sheet, cellRef(3, r), row.DescEN)
		setCell(f, sheet, cellRef(4, r), row.Qty)
		setCell(f, sheet, cellRef(5, r), row.Pkgs)
		setCell(f, sheet, cellRef(6, r), row.TotalNW)
		setCell(f, sheet, cellRef(7, r), gwValues[i])
		setCell(f, sheet, cellRef(8, r), row.TotalGW)
		setCell(f, sheet, cellRef(9, r), row.PkgDim)
		setCell(f, sheet, cellRef(10, r), row.Volume)
		setCell(f, sheet, cellRef(11, r), row.PkgMark)
		setCell(f, sheet, cellRef(12, r), row.Manufacturer)
		setCell(f, sheet, cellRef(13, r), row.TradeMark)
	}

	totalRow = PLDataRow + xx
	setCell(f, sheet, cellRef(4, totalRow), round2(sumColumn(rows, func(r DataRecord) float64 { return r.Qty })))
	setCell(f, sheet, cellRef(5, totalRow), round2(sumColumn(rows, func(r DataRecord) float64 { return r.Pkgs })))
	setCell(f, sheet, cellRef(6, totalRow), round2(sumColumn(rows, func(r DataRecord) float64 { return r.TotalNW })))
	setCell(f, sheet, cellRef(7, totalRow), round2(sumFloats(gwValues)))
	setCell(f, sheet, cellRef(8, totalRow), plTotalGW(rows))
	setCell(f, sheet, cellRef(10, totalRow), round2(sumColumn(rows, func(r DataRecord) float64 { return r.Volume })))
	return nil
}

// calcPLGWValues 末行 GW 按总毛重倒减，保证合计一致
func calcPLGWValues(rows []DataRecord) []float64 {
	out := make([]float64, len(rows))
	var sum float64
	target := rows[0].TotalGW
	if target <= 0 {
		for i, row := range rows {
			out[i] = round2(row.GW)
		}
		return out
	}
	for i := 0; i < len(rows)-1; i++ {
		out[i] = round2(rows[i].GW)
		sum += out[i]
	}
	out[len(rows)-1] = round2(target - sum)
	if out[len(rows)-1] < 0 {
		out[len(rows)-1] = round2(rows[len(rows)-1].GW)
	}
	return out
}

func plTotalGW(rows []DataRecord) float64 {
	if len(rows) == 0 {
		return 0
	}
	if rows[0].TotalGW > 0 {
		return round2(rows[0].TotalGW)
	}
	return round2(sumColumn(rows, func(r DataRecord) float64 { return r.GW }))
}

func prepareDataRows(f *excelize.File, sheet string, dataRow, insertRow, count int) error {
	insertCount := count - 1
	if insertCount <= 0 {
		return nil
	}
	if err := f.InsertRows(sheet, insertRow, insertCount); err != nil {
		return fmt.Errorf("插入行失败: %w", err)
	}
	copyRowStyle(f, sheet, dataRow, insertRow, insertCount, 9)
	return nil
}

func findTotalRow(f *excelize.File, sheet string, startRow int) int {
	for r := startRow; r <= startRow+500; r++ {
		v, _ := f.GetCellValue(sheet, cellRef(1, r))
		if strings.Contains(strings.ToUpper(strings.TrimSpace(v)), "TOTAL") {
			return r
		}
	}
	return 0
}

func getCellStr(f *excelize.File, sheet string, row, col int) string {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	v, _ := f.GetCellValue(sheet, cell)
	return strings.TrimSpace(v)
}

func getCellFloat(f *excelize.File, sheet string, row, col int) float64 {
	s := getCellStr(f, sheet, row, col)
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func resolveSheetName(f *excelize.File, target string) string {
	for _, name := range f.GetSheetList() {
		if name == target {
			return name
		}
	}
	for _, name := range f.GetSheetList() {
		if strings.EqualFold(strings.TrimSpace(name), target) {
			return name
		}
	}
	return ""
}

func UniqueHSCodesSorted(records []DataRecord) []string {
	seen := make(map[string]bool)
	var codes []string
	for _, r := range records {
		if !seen[r.HSCode] {
			seen[r.HSCode] = true
			codes = append(codes, r.HSCode)
		}
	}
	sort.Strings(codes)
	return codes
}

func FilterByHSCode(records []DataRecord, hsCode string) []DataRecord {
	var result []DataRecord
	for _, r := range records {
		if r.HSCode == hsCode {
			result = append(result, r)
		}
	}
	return result
}

func GetLastHSCodeFromOutputDir(outputDir string) string {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return ""
	}
	var lastHS string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if ext != ".xlsx" && ext != ".xlsm" {
			continue
		}
		base := strings.TrimSuffix(name, ext)
		m := hsCodeFilePattern.FindStringSubmatch(base)
		if len(m) >= 3 {
			hs := m[2]
			if lastHS == "" || hs > lastHS {
				lastHS = hs
			}
		}
	}
	return lastHS
}

func NextHSCode(allCodes []string, lastGenerated string) string {
	if lastGenerated == "" {
		if len(allCodes) > 0 {
			return allCodes[0]
		}
		return ""
	}
	for i, code := range allCodes {
		if code == lastGenerated && i+1 < len(allCodes) {
			return allCodes[i+1]
		}
	}
	return ""
}

func GenerateCustomsExcel(templatePath, outputDir, hsCode string, rows []DataRecord) (*GenerateResult, error) {
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("读取模板失败: %w", err)
	}
	return GenerateFromTemplateBytes(templateData, outputDir, hsCode, rows)
}

func writeFileWithRetry(path string, data []byte) error {
	var lastErr error
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(path, data, 0644); err != nil {
			lastErr = err
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

func RealTemplatePath() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 6; i++ {
		p := filepath.Join(dir, "ATP海运模板小程序.xlsm")
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			if abs != "" {
				return abs
			}
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func copyRowStyle(f *excelize.File, sheet string, srcRow, startRow, count, maxCol int) {
	for i := 0; i < count; i++ {
		targetRow := startRow + i
		for col := 1; col <= maxCol; col++ {
			srcCell := cellRef(col, srcRow)
			dstCell := cellRef(col, targetRow)
			styleID, err := f.GetCellStyle(sheet, srcCell)
			if err == nil && styleID > 0 {
				_ = f.SetCellStyle(sheet, dstCell, dstCell, styleID)
			}
		}
	}
}

func setCell(f *excelize.File, sheet, cell string, value interface{}) {
	_ = f.SetCellFormula(sheet, cell, "")
	switch v := value.(type) {
	case float64:
		_ = f.SetCellValue(sheet, cell, v)
	case int:
		_ = f.SetCellValue(sheet, cell, v)
	default:
		_ = f.SetCellValue(sheet, cell, value)
	}
}

func cellRef(col, row int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return name + strconv.Itoa(row)
}

func colLetter(col int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return name
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func sumColumn(rows []DataRecord, getter func(DataRecord) float64) float64 {
	var sum float64
	for _, r := range rows {
		sum += getter(r)
	}
	return sum
}

func sumFloats(vals []float64) float64 {
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum
}

func sanitizeFileName(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		"\\", "_", "/", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(s)
}

func ValidateSourceFile(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %w", err)
	}
	defer f.Close()
	if resolveSheetName(f, SheetData) == "" {
		return fmt.Errorf("源文件缺少「数据」工作表")
	}
	return nil
}

func ValidateTemplateFile(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("无法打开模板文件: %w", err)
	}
	defer f.Close()
	if resolveSheetName(f, SheetInvoice) == "" {
		return fmt.Errorf("模板文件缺少「INVOICE」工作表")
	}
	if resolveSheetName(f, SheetPL) == "" {
		return fmt.Errorf("模板文件缺少「PL」工作表")
	}
	return nil
}

func BatchDirName(t time.Time) string {
	return t.Format("20060102_150405")
}

func ZipFileName(batchName string) string {
	return batchName + "_customs.zip"
}
