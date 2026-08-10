// Package model defines the input data structures for the Squash TM test
// case generator. These types mirror the YAML file structure that the user
// (or Claude) writes by hand for each user story / feature.
package model

// Suite represents one Squash TM folder worth of test cases (typically
// one user story, e.g. "EX-001 - Filter items list").
type Suite struct {
	// Project is the Squash TM project name, e.g. "My Project".
	Project string `yaml:"project"`

	// FolderPath is the full path of the target folder in the Test Case
	// workspace, e.g.:
	// "/My Project/My Feature/01 - My organization's Folder/EX-001 - Filter items list"
	FolderPath string `yaml:"folder_path"`

	// Reference is written into TC_REFERENCE for every test case in this
	// suite (Squash TM convention observed: the story ID, not a per-CT id).
	Reference string `yaml:"reference"`

	// RequirementPath, if set, is used to populate LINK_REQ_TC so every
	// test case in this suite is automatically linked to the requirement.
	RequirementPath string `yaml:"requirement_path,omitempty"`

	TestCases []TestCase `yaml:"test_cases"`
}

// TestCase represents a single Squash TM classic test case.
type TestCase struct {
	// Title becomes the test case name, conventionally "[BR-x] ...".
	Title string `yaml:"title"`

	// Weight is Squash TM's importance field (HIGH / MEDIUM / LOW).
	Weight string `yaml:"weight"`

	// Prereq is the prerequisite text (plain text; the generator wraps it
	// in the minimal HTML Squash TM's rich text field expects).
	Prereq string `yaml:"prereq"`

	Steps []Step `yaml:"steps"`

	// Parameters and Datasets are optional: fill them in only when several
	// Business Rules share the exact same test logic and only differ by a
	// value (e.g. BR-4/5/6/7 differing only by which capability attribute
	// is searched).
	Parameters []string  `yaml:"parameters,omitempty"`
	Datasets   []Dataset `yaml:"datasets,omitempty"`
}

// Step is one action/expected-result pair. Use {PARAM_NAME} placeholders in
// Action/Expected to reference a Parameter defined on the parent TestCase.
type Step struct {
	Action   string `yaml:"action"`
	Expected string `yaml:"expected"`
}

// Dataset is one set of concrete values for a parameterized test case.
// Values keys must match the TestCase.Parameters names exactly.
type Dataset struct {
	Name   string            `yaml:"name"`
	Values map[string]string `yaml:"values"`
}
