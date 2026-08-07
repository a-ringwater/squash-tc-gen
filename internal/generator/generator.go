// Package generator builds a Squash TM-compatible import .xlsx file
// (TEST_CASES / STEPS / PARAMETERS / DATASETS / LINK_REQ_TC tabs) from a
// model.Suite. Column headers are the REAL ones confirmed by exporting an
// existing folder from the target Squash TM instance (Gazelle v2) on
// 2026-08-07 — do not "guess-fix" them without re-checking a fresh export.
package generator

import (
	"fmt"
	"html"

	"github.com/xuri/excelize/v2"

	"github.com/adrien/squash-tc-gen/internal/model"
)

// Confirmed real headers, in the exact order Squash TM expects.
var (
	tcHeaders = []string{
		"ACTION", "PROJECT_ID", "PROJECT_NAME", "TC_PATH", "TC_NUM", "TC_ID", "TC_UUID",
		"TC_REFERENCE", "TC_NAME", "TC_MILESTONE", "TC_WEIGHT_AUTO", "TC_WEIGHT", "TC_NATURE",
		"TC_TYPE", "TC_STATUS", "TC_DESCRIPTION", "TC_PRE_REQUISITE", "TC_#_REQ",
		"TC_#_CALLED_BY", "TC_#_ATTACHMENT", "TC_CREATED_ON", "TC_CREATED_BY",
		"TC_LAST_MODIFIED_ON", "TC_LAST_MODIFIED_BY", "TC_KIND", "TC_SCRIPT", "TC_AUTOMATABLE",
	}
	stepHeaders = []string{
		"ACTION", "TC_OWNER_PATH", "TC_OWNER_ID", "TC_STEP_ID", "TC_STEP_NUM",
		"TC_STEP_IS_CALL_STEP", "TC_STEP_CALL_DATASET", "TC_STEP_ACTION",
		"TC_STEP_EXPECTED_RESULT", "TC_STEP_#_REQ", "TC_STEP_#_ATTACHMENT",
	}
	paramHeaders = []string{
		"ACTION", "TC_OWNER_PATH", "TC_OWNER_ID", "TC_PARAM_ID", "TC_PARAM_NAME", "TC_PARAM_DESCRIPTION",
	}
	datasetHeaders = []string{
		"ACTION", "TC_OWNER_PATH", "TC_OWNER_ID", "TC_DATASET_ID", "TC_DATASET_NAME",
		"TC_PARAM_OWNER_PATH", "TC_PARAM_OWNER_ID", "TC_DATASET_PARAM_NAME", "TC_DATASET_PARAM_VALUE",
	}
	linkHeaders = []string{"REQ_PATH", "REQ_VERSION_NUM", "TC_PATH"}
)

// richText wraps plain text into the minimal HTML Squash TM's rich text
// fields (prerequisite, step action/expected result) expect, escaping any
// HTML-sensitive characters in the source text.
func richText(s string) string {
	return fmt.Sprintf(`<p><span style="font-family:Arial">%s</span></p>`, html.EscapeString(s))
}

// headerStyle / bodyStyle are created once per file and reused for every cell.
type styles struct {
	header int
	body   int
}

func newStyles(f *excelize.File) (styles, error) {
	header, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Family: "Arial", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4E78"}, Pattern: 1},
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top", Horizontal: "center"},
	})
	if err != nil {
		return styles{}, err
	}
	body, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Arial", Size: 10},
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"},
	})
	if err != nil {
		return styles{}, err
	}
	return styles{header: header, body: body}, nil
}

func writeHeaderRow(f *excelize.File, sheet string, headers []string, st styles) error {
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
	}
	endCell, _ := excelize.CoordinatesToCellName(len(headers), 1)
	if err := f.SetCellStyle(sheet, "A1", endCell, st.header); err != nil {
		return err
	}
	return f.SetRowHeight(sheet, 1, 26)
}

func writeRow(f *excelize.File, sheet string, row int, values []interface{}, st styles) error {
	for i, v := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			return err
		}
	}
	endCell, _ := excelize.CoordinatesToCellName(len(values), row)
	startCell, _ := excelize.CoordinatesToCellName(1, row)
	return f.SetCellStyle(sheet, startCell, endCell, st.body)
}

// Generate builds the Squash TM import workbook for the given suite and
// writes it to outputPath.
func Generate(suite model.Suite, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	st, err := newStyles(f)
	if err != nil {
		return fmt.Errorf("creating styles: %w", err)
	}

	// excelize creates a default "Sheet1" — rename it to our first tab.
	f.SetSheetName("Sheet1", "TEST_CASES")
	for _, name := range []string{"STEPS", "PARAMETERS", "DATASETS", "LINK_REQ_TC"} {
		if _, err := f.NewSheet(name); err != nil {
			return fmt.Errorf("creating sheet %s: %w", name, err)
		}
	}

	if err := writeHeaderRow(f, "TEST_CASES", tcHeaders, st); err != nil {
		return err
	}
	if err := writeHeaderRow(f, "STEPS", stepHeaders, st); err != nil {
		return err
	}
	if err := writeHeaderRow(f, "PARAMETERS", paramHeaders, st); err != nil {
		return err
	}
	if err := writeHeaderRow(f, "DATASETS", datasetHeaders, st); err != nil {
		return err
	}
	if err := writeHeaderRow(f, "LINK_REQ_TC", linkHeaders, st); err != nil {
		return err
	}

	tcRow, stepRow, paramRow, datasetRow, linkRow := 2, 2, 2, 2, 2

	for _, tc := range suite.TestCases {
		fullPath := suite.FolderPath + "/" + tc.Title

		// --- TEST_CASES ---
		tcValues := []interface{}{
			"C", "", "", fullPath, "", "", "",
			suite.Reference, tc.Title, "", "", tc.Weight, "",
			"", "WORK_IN_PROGRESS", "", richText(tc.Prereq), "",
			"", "", "", "",
			"", "", "STANDARD", "", "",
		}
		if err := writeRow(f, "TEST_CASES", tcRow, tcValues, st); err != nil {
			return err
		}
		tcRow++

		// --- STEPS ---
		for i, step := range tc.Steps {
			stepValues := []interface{}{
				"C", fullPath, "", "", i + 1,
				0, "", richText(step.Action), richText(step.Expected), "", "",
			}
			if err := writeRow(f, "STEPS", stepRow, stepValues, st); err != nil {
				return err
			}
			stepRow++
		}

		// --- PARAMETERS ---
		for _, p := range tc.Parameters {
			paramValues := []interface{}{"C", fullPath, "", "", p, ""}
			if err := writeRow(f, "PARAMETERS", paramRow, paramValues, st); err != nil {
				return err
			}
			paramRow++
		}

		// --- DATASETS ---
		for _, ds := range tc.Datasets {
			for _, p := range tc.Parameters {
				val := ds.Values[p]
				dsValues := []interface{}{
					"C", fullPath, "", "", ds.Name,
					fullPath, "", p, val,
				}
				if err := writeRow(f, "DATASETS", datasetRow, dsValues, st); err != nil {
					return err
				}
				datasetRow++
			}
		}

		// --- LINK_REQ_TC (only if a requirement path was given) ---
		if suite.RequirementPath != "" {
			linkValues := []interface{}{suite.RequirementPath, 1, fullPath}
			if err := writeRow(f, "LINK_REQ_TC", linkRow, linkValues, st); err != nil {
				return err
			}
			linkRow++
		}
	}

	applyColumnWidths(f)

	f.SetActiveSheet(0)
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("saving %s: %w", outputPath, err)
	}
	return nil
}

func applyColumnWidths(f *excelize.File) {
	tcWidths := map[string]float64{
		"A": 8, "B": 10, "C": 12, "D": 75, "E": 8, "F": 8, "G": 8, "H": 12, "I": 45,
		"J": 14, "K": 12, "L": 10, "M": 12, "N": 12, "O": 16, "P": 14, "Q": 40, "R": 8,
		"S": 12, "T": 12, "U": 12, "V": 12, "W": 14, "X": 14, "Y": 10, "Z": 10, "AA": 14,
	}
	for col, w := range tcWidths {
		f.SetColWidth("TEST_CASES", col, col, w)
	}

	stepWidths := map[string]float64{
		"A": 8, "B": 75, "C": 10, "D": 10, "E": 10, "F": 16, "G": 16, "H": 45, "I": 45, "J": 8, "K": 10,
	}
	for col, w := range stepWidths {
		f.SetColWidth("STEPS", col, col, w)
	}

	paramWidths := map[string]float64{"A": 8, "B": 75, "C": 10, "D": 10, "E": 20, "F": 25}
	for col, w := range paramWidths {
		f.SetColWidth("PARAMETERS", col, col, w)
	}

	datasetWidths := map[string]float64{
		"A": 8, "B": 75, "C": 10, "D": 10, "E": 16, "F": 75, "G": 10, "H": 18, "I": 45,
	}
	for col, w := range datasetWidths {
		f.SetColWidth("DATASETS", col, col, w)
	}

	linkWidths := map[string]float64{"A": 75, "B": 14, "C": 75}
	for col, w := range linkWidths {
		f.SetColWidth("LINK_REQ_TC", col, col, w)
	}
}
