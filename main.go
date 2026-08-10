// squash-tc-gen reads a YAML file describing a user story's test cases and
// writes a Squash TM-compatible Excel import file.
//
// Usage:
//
//	squash-tc-gen -in examples/example.yaml -out import.xlsx
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/adrien/squash-tc-gen/internal/generator"
	"github.com/adrien/squash-tc-gen/internal/model"
)

func main() {
	// flag.String returns a *string, not a string: flag.Parse() below fills
	// it in later once os.Args has actually been read, so at declaration
	// time we only get the address of where the value will end up.
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

// run holds the actual program logic, kept separate from main so it stays
// testable: main calls os.Exit, which a test cannot easily intercept.
func run(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		// %w (not %v) wraps err instead of just stringifying it, so a
		// caller could still errors.Is/As into the original error.
		return fmt.Errorf("reading %s: %w", inputPath, err)
	}

	var suite model.Suite
	// &suite: Unmarshal must write into our variable, so it needs a
	// pointer to it rather than a copy.
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

// validWeights is the set of accepted TestCase.Weight values, used as a
// map[string]bool "set": a missing key reads as its zero value (false),
// so validWeights[x] alone answers "is x in the set?". Declared once at
// package level instead of rebuilt on every validate call.
var validWeights = map[string]bool{"HIGH": true, "MEDIUM": true, "LOW": true}

// validateDatasets checks that every dataset provides a value for each
// parameter declared on tc. Split out of validate so its double loop keeps
// its own nesting level instead of adding to validate's.
func validateDatasets(tc model.TestCase) error {
	for _, ds := range tc.Datasets {
		for _, p := range tc.Parameters {
			// comma-ok form: distinguishes "key present with an empty
			// value" from "key absent entirely" — a plain ds.Values[p]
			// would return "" for both cases.
			if _, ok := ds.Values[p]; !ok {
				return fmt.Errorf("dataset %q is missing a value for parameter %q", ds.Name, p)
			}
		}
	}
	return nil
}

// validate does a few sanity checks so mistakes in the YAML fail fast with a
// clear message instead of producing a silently broken import file.
func validate(s model.Suite) error {
	if s.Project == "" {
		return errors.New("project is required")
	}
	if s.FolderPath == "" {
		return errors.New("folder_path is required")
	}
	if len(s.TestCases) == 0 {
		return errors.New("at least one test case is required")
	}
	for _, tc := range s.TestCases {
		if tc.Title == "" {
			return errors.New("a test case has no title")
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
		if err := validateDatasets(tc); err != nil {
			// validateDatasets' own error is generic ("dataset %q is
			// missing..."); wrap it here to add which test case it came
			// from, instead of passing tc.Title into that function.
			return fmt.Errorf("test case %q: %w", tc.Title, err)
		}
	}
	return nil
}
