package importjob

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

var ErrUnsupportedFile = errors.New("only csv, xlsx and xlsm files are supported")

var headerAliases = map[string]string{
	"slug": "slug", "标识": "slug", "latin_name": "latin_name", "拉丁名": "latin_name", "chinese_name": "chinese_name", "中文名": "chinese_name",
	"strain_number": "strain_number", "菌株编号": "strain_number", "source_environment": "source_environment", "来源环境": "source_environment",
	"safety_level": "safety_level", "安全等级": "safety_level", "is_model_organism": "is_model_organism", "模式菌": "is_model_organism",
	"summary": "summary", "摘要": "summary", "function_tags": "function_tags", "功能标签": "function_tags", "medium_name": "medium_name", "培养基": "medium_name",
	"aliases": "aliases", "别名": "aliases", "同义词": "aliases",
	"temperature_min": "temperature_min", "最低温度": "temperature_min", "temperature_max": "temperature_max", "最高温度": "temperature_max",
	"ph_min": "ph_min", "最低ph": "ph_min", "ph_max": "ph_max", "最高ph": "ph_max", "oxygen_requirement": "oxygen_requirement", "氧需求": "oxygen_requirement",
	"culture_time": "culture_time", "培养时间": "culture_time",
}

func Parse(filename string, reader io.Reader) ([]SpeciesRow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var records [][]string
	var err error
	switch ext {
	case ".csv":
		csvReader := csv.NewReader(reader)
		csvReader.TrimLeadingSpace = true
		csvReader.FieldsPerRecord = -1
		records, err = csvReader.ReadAll()
	case ".xlsx", ".xlsm":
		file, openErr := excelize.OpenReader(reader)
		if openErr != nil {
			return nil, openErr
		}
		defer file.Close()
		sheets := file.GetSheetList()
		if len(sheets) == 0 {
			return nil, errors.New("workbook has no worksheet")
		}
		records, err = file.GetRows(sheets[0])
	default:
		return nil, ErrUnsupportedFile
	}
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, errors.New("file must contain a header and at least one data row")
	}
	headers := make([]string, len(records[0]))
	foundHeaders := map[string]bool{}
	for index, value := range records[0] {
		headers[index] = headerAliases[strings.ToLower(strings.TrimSpace(value))]
		foundHeaders[headers[index]] = true
	}
	if !foundHeaders["slug"] || !foundHeaders["latin_name"] {
		return nil, errors.New("header must contain slug and latin_name (or 拉丁名)")
	}
	rows := make([]SpeciesRow, 0, len(records)-1)
	for index, record := range records[1:] {
		values := map[string]string{}
		for column, value := range record {
			if column < len(headers) && headers[column] != "" {
				values[headers[column]] = strings.TrimSpace(value)
			}
		}
		if allEmpty(values) {
			continue
		}
		row := SpeciesRow{RowNumber: index + 2, Slug: values["slug"], LatinName: values["latin_name"], ChineseName: values["chinese_name"], StrainNumber: values["strain_number"], SourceEnvironment: values["source_environment"], SafetyLevel: values["safety_level"], Summary: values["summary"], MediumName: values["medium_name"], OxygenRequirement: values["oxygen_requirement"], CultureTime: values["culture_time"]}
		row.IsModelOrganism = parseBool(values["is_model_organism"])
		if value := values["function_tags"]; value != "" {
			for _, tag := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' || r == '，' || r == '；' }) {
				if tag = strings.TrimSpace(tag); tag != "" {
					row.FunctionTags = append(row.FunctionTags, tag)
				}
			}
		}
		if value := values["aliases"]; value != "" {
			for _, alias := range strings.FieldsFunc(value, func(r rune) bool { return r == ';' || r == '；' || r == '|' }) {
				if alias = strings.TrimSpace(alias); alias != "" {
					row.Aliases = append(row.Aliases, alias)
				}
			}
		}
		row.TemperatureMin = parseNumber(values["temperature_min"], "最低温度", &row.Errors)
		row.TemperatureMax = parseNumber(values["temperature_max"], "最高温度", &row.Errors)
		row.PHMin = parseNumber(values["ph_min"], "最低 pH", &row.Errors)
		row.PHMax = parseNumber(values["ph_max"], "最高 pH", &row.Errors)
		validate(&row)
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, errors.New("file contains no data rows")
	}
	return rows, nil
}

func allEmpty(values map[string]string) bool {
	for _, value := range values {
		if value != "" {
			return false
		}
	}
	return true
}
func parseBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes" || value == "是"
}
func parseNumber(value, label string, errs *[]string) *float64 {
	if value == "" {
		return nil
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		*errs = append(*errs, fmt.Sprintf("%s不是有效数字", label))
		return nil
	}
	return &number
}
func validate(row *SpeciesRow) {
	if row.Slug == "" {
		row.Errors = append(row.Errors, "slug 不能为空")
	}
	if row.LatinName == "" {
		row.Errors = append(row.Errors, "拉丁名不能为空")
	}
	if strings.ContainsAny(row.Slug, " /\\") {
		row.Errors = append(row.Errors, "slug 不能包含空格或斜杠")
	}
	if row.PHMin != nil && (*row.PHMin < 0 || *row.PHMin > 14) || row.PHMax != nil && (*row.PHMax < 0 || *row.PHMax > 14) {
		row.Errors = append(row.Errors, "pH 必须在 0–14 之间")
	}
	if row.TemperatureMin != nil && row.TemperatureMax != nil && *row.TemperatureMin > *row.TemperatureMax {
		row.Errors = append(row.Errors, "最低温度不能高于最高温度")
	}
	if row.PHMin != nil && row.PHMax != nil && *row.PHMin > *row.PHMax {
		row.Errors = append(row.Errors, "最低 pH 不能高于最高 pH")
	}
}
