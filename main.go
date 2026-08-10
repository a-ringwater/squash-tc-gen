// squash-tc-gen reads a YAML file describing a user story's test cases and
// writes a Squash TM-compatible Excel import file.
//
// Usage:
//
//	squash-tc-gen -in examples/example.yaml -out import.xlsx
package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/adrien/squash-tc-gen/internal/generator"
	"github.com/adrien/squash-tc-gen/internal/model"
)

func main() {
	inputPath := flag.String("in", "", "path to the input YAML file (required)")
	outputPath := flag.String("out", "import.xlsx", "path to the generated .xlsx file")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "error: -in is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*inputPath, *outputPath); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", inputPath, err)
	}

	var suite model.Suite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return fmt.Errorf("parsing YAML: %w", err)
	}

	if err := validate(suite); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	if err := generator.Generate(suite, outputPath); err != nil {
		return fmt.Errorf("generating xlsx: %w", err)
	}

	fmt.Printf("OK: %d test case(s) written to %s\n", len(suite.TestCases), outputPath)
	return nil
}

// validate does a few sanity checks so mistakes in the YAML fail fast with a
// clear message instead of producing a silently broken import file.
func validate(s model.Suite) error {
	if s.Project == "" {
		return fmt.Errorf("project is required")
	}
	if s.FolderPath == "" {
		return fmt.Errorf("folder_path is required")
	}
	if len(s.TestCases) == 0 {
		return fmt.Errorf("at least one test case is required")
	}
	validWeights := map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}
	for _, tc := range s.TestCases {
		if tc.Title == "" {
			return fmt.Errorf("a test case has no title")
		}
		if !validWeights[tc.Weight] {
			return fmt.Errorf("test case %q: weight must be HIGH, MEDIUM or LOW (got %q)", tc.Title, tc.Weight)
		}
		if len(tc.Steps) == 0 {
			return fmt.Errorf("test case %q has no steps", tc.Title)
		}
		if len(tc.Datasets) > 0 && len(tc.Parameters) == 0 {
			return fmt.Errorf("test case %q has datasets but no parameters declared", tc.Title)
		}
		for _, ds := range tc.Datasets {
			for _, p := range tc.Parameters {
				if _, ok := ds.Values[p]; !ok {
					return fmt.Errorf("test case %q: dataset %q is missing a value for parameter %q", tc.Title, ds.Name, p)
				}
			}
		}
	}
	return nil
}
