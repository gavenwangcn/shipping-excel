package excel

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestRealTemplateParse(t *testing.T) {
	path := RealTemplatePath()
	if path == "" {
		t.Skip("未找到 ATP海运模板小程序.xlsm，跳过真实文件测试")
	}

	records, err := ParseDataSheet(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1186 {
		t.Fatalf("期望 1186 条数据，实际 %d", len(records))
	}

	codes := UniqueHSCodesSorted(records)
	if len(codes) != 43 {
		t.Fatalf("期望 43 个 HS CODE，实际 %d", len(codes))
	}

	first := records[0]
	if first.HSCode == "" || first.PartNo == "" {
		t.Fatal("首条记录字段为空")
	}
	if first.CINo != "43115-DT1EJ20250101-A" && first.CINo != "43115-DT1EJ20250101-B" {
		t.Logf("CI NO sample: %s", first.CINo)
	}
}

func TestRealTemplateGenerate3926300000(t *testing.T) {
	path := RealTemplatePath()
	if path == "" {
		t.Skip("未找到 ATP海运模板小程序.xlsm，跳过真实文件测试")
	}

	records, err := ParseDataSheet(path)
	if err != nil {
		t.Fatal(err)
	}

	hs := "3926300000"
	filtered := FilterByHSCode(records, hs)
	if len(filtered) != 4 {
		t.Fatalf("HS %s 期望 4 行，实际 %d", hs, len(filtered))
	}

	templateData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join("testdata", "real-test-output")
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)
	t.Cleanup(func() { os.RemoveAll(outDir) })

	result, err := GenerateFromTemplateBytes(templateData, outDir, hs, filtered)
	if err != nil {
		t.Fatal(err)
	}

	wantName := "43115-DT1EJ20250101-A" + hs + ".xlsx"
	if result.FileName != wantName {
		t.Fatalf("文件名期望 %s，实际 %s", wantName, result.FileName)
	}

	f, err := excelize.OpenFile(result.FilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	inv := SheetInvoice
	for i, row := range filtered {
		r := InvoiceDataRow + i
		a, _ := f.GetCellValue(inv, "A"+strconv.Itoa(r))
		if a != strconv.Itoa(i+1) {
			t.Fatalf("R%d 序号错误: %s", r, a)
		}
		b, _ := f.GetCellValue(inv, "B"+strconv.Itoa(r))
		if b != row.PartNo {
			t.Fatalf("R%d PART NO 错误: %s vs %s", r, b, row.PartNo)
		}
		c, _ := f.GetCellValue(inv, "C"+strconv.Itoa(r))
		if c != row.DescEN {
			t.Fatalf("R%d 英文描述不匹配", r)
		}
		h, _ := f.GetCellValue(inv, "H"+strconv.Itoa(r))
		if h != row.DescRU {
			t.Fatalf("R%d 俄文描述不匹配", r)
		}
		wantG := round2(row.Qty * row.UnitPrice)
		gStr := strings.TrimSpace(getCell(t, f, inv, "G"+strconv.Itoa(r)))
		eStr := strings.TrimSpace(getCell(t, f, inv, "E"+strconv.Itoa(r)))
		fStr := strings.TrimSpace(getCell(t, f, inv, "F"+strconv.Itoa(r)))
		g, err := strconv.ParseFloat(gStr, 64)
		if err != nil || round2(g) != wantG {
			t.Fatalf("R%d 金额错误: E=%q F=%q G=%q parsed=%.2f want %.2f", r, eStr, fStr, gStr, g, wantG)
		}
	}

	xx := len(filtered)
	freightRow := InvoiceDataRow + xx
	insRow := freightRow + 1
	totalRow := insRow + 1

	aFreight, _ := f.GetCellValue(inv, "A"+strconv.Itoa(freightRow))
	if aFreight != "Freight Value(CNY)" {
		t.Fatalf("运费行标签错误: %q", aFreight)
	}

	var sumQ, sumR, sumG float64
	for _, row := range filtered {
		sumQ += row.Freight1
		sumR += row.Insurance
		sumG += row.Qty * row.UnitPrice
	}

	assertCellFloat(t, f, inv, "G"+strconv.Itoa(freightRow), sumQ)
	assertCellFloat(t, f, inv, "G"+strconv.Itoa(insRow), sumR)
	assertCellFloat(t, f, inv, "G"+strconv.Itoa(totalRow), sumQ+sumR+sumG)

	// 模板头部应保留
	title, _ := f.GetCellValue(inv, "A3")
	if !strings.Contains(title, "COMMERCIAL INVOICE") {
		t.Fatalf("INVOICE 模板头部丢失: %q", title)
	}
	i13, _ := f.GetCellValue(inv, "I13")
	if i13 != hs {
		t.Fatalf("INVOICE I13 HS CODE 错误: %q", i13)
	}

	// PL 表数据区与汇总
	pl := SheetPL
	plTitle, _ := f.GetCellValue(pl, "A3")
	if !strings.Contains(plTitle, "Packing List") {
		t.Fatalf("PL 模板头部丢失: %q", plTitle)
	}
	b13, _ := f.GetCellValue(pl, "B13")
	if b13 != filtered[0].PartNo {
		t.Fatalf("PL B13 错误: %q", b13)
	}
	plTotalRow := PLDataRow + xx
	aPlTotal, _ := f.GetCellValue(pl, "A"+strconv.Itoa(plTotalRow))
	if !strings.Contains(strings.ToUpper(aPlTotal), "TOTAL") {
		t.Fatalf("PL TOTAL 行位置错误: row=%d val=%q", plTotalRow, aPlTotal)
	}

	zipPath := filepath.Join(outDir, "test.zip")
	if err := CreateZipFromDir(outDir, zipPath); err != nil {
		t.Fatal(err)
	}
}

func TestRealTemplateGenerateAllHSCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过全量生成测试 (short mode)")
	}
	if os.Getenv("CI") == "" && filepath.Separator == '\\' {
		t.Skip("Windows 下跳过多文件批量生成测试（防病毒软件文件锁），请在 Linux/Docker 中运行")
	}

	path := RealTemplatePath()
	if path == "" {
		t.Skip("未找到 ATP海运模板小程序.xlsm，跳过真实文件测试")
	}

	records, err := ParseDataSheet(path)
	if err != nil {
		t.Fatal(err)
	}
	templateData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(t.TempDir(), "output")
	codes := UniqueHSCodesSorted(records)

	for _, hs := range codes {
		filtered := FilterByHSCode(records, hs)
		_, err := GenerateFromTemplateBytes(templateData, outDir, hs, filtered)
		if err != nil {
			t.Fatalf("生成 HS %s 失败: %v", hs, err)
		}
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 43 {
		t.Fatalf("期望 43 个输出文件，实际 %d", len(entries))
	}
}

func getCell(t *testing.T, f *excelize.File, sheet, cell string) string {
	t.Helper()
	v, err := f.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("read %s: %v", cell, err)
	}
	return v
}

func assertCellFloat(t *testing.T, f *excelize.File, sheet, cell string, want float64) {
	t.Helper()
	v := strings.TrimSpace(getCell(t, f, sheet, cell))
	got, err := strconv.ParseFloat(v, 64)
	if err != nil {
		t.Fatalf("cell %s 非数字: %q", cell, v)
	}
	if round2(got) != round2(want) {
		t.Fatalf("cell %s: got %.2f want %.2f", cell, got, want)
	}
}
