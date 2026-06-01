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
)

type DataRecord struct {
	RowNum    int
	CINo      string
	PartNo    string
	HSCode    string
	DescEN    string
	DescRU    string
	Qty       float64
	UnitPrice float64
	Freight1  float64
	Insurance float64
	Type      string
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
		hsCode := getCellStr(f, sheetName, i, 12) // L
		if hsCode == "" {
			continue
		}
		records = append(records, DataRecord{
			RowNum:    i,
			CINo:      getCellStr(f, sheetName, i, 3),  // C
			PartNo:    getCellStr(f, sheetName, i, 11), // K
			HSCode:    hsCode,
			DescEN:    getCellStr(f, sheetName, i, 13), // M
			DescRU:    getCellStr(f, sheetName, i, 14), // N
			Qty:       getCellFloat(f, sheetName, i, 15),
			UnitPrice: getCellFloat(f, sheetName, i, 16),
			Freight1:  getCellFloat(f, sheetName, i, 17),
			Insurance: getCellFloat(f, sheetName, i, 18),
			Type:      getCellStr(f, sheetName, i, 30), // AD
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].HSCode < records[j].HSCode
	})

	return records, nil
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
	if len(rows) == 0 {
		return nil, fmt.Errorf("HS CODE %s 无数据行", hsCode)
	}
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("读取模板失败: %w", err)
	}
	return generateFromTemplateBytes(templateData, outputDir, hsCode, rows)
}

func GenerateFromTemplateBytes(templateData []byte, outputDir, hsCode string, rows []DataRecord) (*GenerateResult, error) {
	return generateFromTemplateBytes(templateData, outputDir, hsCode, rows)
}

func generateFromTemplateBytes(templateData []byte, outputDir, hsCode string, rows []DataRecord) (*GenerateResult, error) {
	ciNo := sanitizeFileName(rows[0].CINo)
	fileName := fmt.Sprintf("%s%s.xlsx", ciNo, hsCode)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}
	outPath := filepath.Join(outputDir, fileName)

	f, err := excelize.OpenReader(bytes.NewReader(templateData))
	if err != nil {
		return nil, fmt.Errorf("打开模板失败: %w", err)
	}

	invoiceSheet := resolveSheetName(f, SheetInvoice)
	if invoiceSheet == "" {
		return nil, fmt.Errorf("模板中未找到 INVOICE 表")
	}

	xx := len(rows)
	insertCount := xx - 1
	if insertCount > 0 {
		if err := f.InsertRows(invoiceSheet, InvoiceInsertRow, insertCount); err != nil {
			return nil, fmt.Errorf("插入行失败: %w", err)
		}
		copyRowStyle(f, invoiceSheet, InvoiceDataRow, InvoiceInsertRow, insertCount)
	}

	var sumG float64
	for i, row := range rows {
		r := InvoiceDataRow + i
		seq := i + 1
		amount := row.Qty * row.UnitPrice

		setCell(f, invoiceSheet, colLetter(1)+strconv.Itoa(r), seq)
		setCell(f, invoiceSheet, colLetter(2)+strconv.Itoa(r), row.PartNo)
		setCell(f, invoiceSheet, colLetter(3)+strconv.Itoa(r), row.DescEN)
		setCell(f, invoiceSheet, colLetter(4)+strconv.Itoa(r), row.Type)
		setCell(f, invoiceSheet, colLetter(5)+strconv.Itoa(r), row.Qty)
		setCell(f, invoiceSheet, colLetter(6)+strconv.Itoa(r), row.UnitPrice)
		setCell(f, invoiceSheet, colLetter(7)+strconv.Itoa(r), round2(amount))
		setCell(f, invoiceSheet, colLetter(8)+strconv.Itoa(r), row.DescRU)

		sumG += amount
	}

	summaryRow := InvoiceDataRow + xx
	sumFreight := sumColumn(rows, func(r DataRecord) float64 { return r.Freight1 })
	sumInsurance := sumColumn(rows, func(r DataRecord) float64 { return r.Insurance })

	setCell(f, invoiceSheet, "G"+strconv.Itoa(summaryRow), round2(sumFreight))
	setCell(f, invoiceSheet, "G"+strconv.Itoa(summaryRow+1), round2(sumInsurance))
	setCell(f, invoiceSheet, "G"+strconv.Itoa(summaryRow+2), round2(sumFreight+sumInsurance+sumG))

	_ = os.Remove(outPath)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成Excel缓冲失败: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("关闭工作簿失败: %w", err)
	}
	if err := writeFileWithRetry(outPath, buf.Bytes()); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	return &GenerateResult{
		FileName: fileName,
		FilePath: outPath,
		HSCode:   hsCode,
		CINo:     rows[0].CINo,
		RowCount: xx,
	}, nil
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
			abs, err := filepath.Abs(p)
			if err == nil {
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

func copyRowStyle(f *excelize.File, sheet string, srcRow, startRow, count int) {
	for i := 0; i < count; i++ {
		targetRow := startRow + i
		for col := 1; col <= 8; col++ {
			srcCell := colLetter(col) + strconv.Itoa(srcRow)
			dstCell := colLetter(col) + strconv.Itoa(targetRow)
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
	return nil
}
