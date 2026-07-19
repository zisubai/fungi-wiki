package importjob

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseCSVWithChineseHeaders(t *testing.T) {
	content := "标识,拉丁名,中文名,功能标签,最低温度,最高温度,最低ph,最高ph\n" +
		"bacillus-demo,Bacillus demo,示例菌,biocontrol;促生,20,35,6,8\n"
	rows, err := Parse("species.csv", strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Slug != "bacillus-demo" || len(rows[0].FunctionTags) != 2 || rows[0].TemperatureMax == nil || *rows[0].TemperatureMax != 35 {
		t.Fatalf("unexpected row: %+v", rows)
	}
}

func TestParseReportsRowValidationErrors(t *testing.T) {
	rows, err := Parse("species.csv", strings.NewReader("slug,latin_name,ph_min,ph_max\nbad slug,,15,5\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows[0].Errors) < 3 {
		t.Fatalf("expected validation errors, got %+v", rows[0].Errors)
	}
}

func TestParseXLSX(t *testing.T) {
	file := excelize.NewFile()
	defer file.Close()
	_ = file.SetSheetRow("Sheet1", "A1", &[]any{"slug", "latin_name", "safety_level"})
	_ = file.SetSheetRow("Sheet1", "A2", &[]any{"yeast-demo", "Saccharomyces demo", "BSL-1"})
	buffer, err := file.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := Parse("species.xlsx", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].LatinName != "Saccharomyces demo" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}
