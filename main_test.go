package main

import (
	"testing"

	"github.com/adrien/squash-tc-gen/internal/model"
)

// validSuite returns a minimal Suite that passes validate(), so each test
// case below only needs to override the one field it wants to break.
func validSuite() model.Suite {
	return model.Suite{
		Project:    "P",
		FolderPath: "/P/folder",
		TestCases: []model.TestCase{
			{
				Title:  "[BR-1] a test case",
				Weight: "HIGH",
				Steps:  []model.Step{{Action: "do", Expected: "see"}},
			},
		},
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		suite   func() model.Suite
		wantErr bool
	}{
		{
			name:    "valid minimal suite",
			suite:   validSuite,
			wantErr: false,
		},
		{
			name: "missing project",
			suite: func() model.Suite {
				s := validSuite()
				s.Project = ""
				return s
			},
			wantErr: true,
		},
		{
			name: "missing folder_path",
			suite: func() model.Suite {
				s := validSuite()
				s.FolderPath = ""
				return s
			},
			wantErr: true,
		},
		{
			name: "no test cases",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases = nil
				return s
			},
			wantErr: true,
		},
		{
			name: "missing title",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases[0].Title = ""
				return s
			},
			wantErr: true,
		},
		{
			name: "invalid weight",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases[0].Weight = "URGENT"
				return s
			},
			wantErr: true,
		},
		{
			name: "no steps",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases[0].Steps = nil
				return s
			},
			wantErr: true,
		},
		{
			name: "datasets without parameters declared",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases[0].Datasets = []model.Dataset{{Name: "case A"}}
				return s
			},
			wantErr: true,
		},
		{
			name: "dataset missing a value for a declared parameter",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases[0].Parameters = []string{"MON_PARAM"}
				s.TestCases[0].Datasets = []model.Dataset{
					{Name: "case A", Values: map[string]string{}},
				}
				return s
			},
			wantErr: true,
		},
		{
			name: "valid parameterized suite",
			suite: func() model.Suite {
				s := validSuite()
				s.TestCases[0].Parameters = []string{"MON_PARAM"}
				s.TestCases[0].Datasets = []model.Dataset{
					{Name: "case A", Values: map[string]string{"MON_PARAM": "valeur A"}},
				}
				return s
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.suite())
			if (err != nil) != tc.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateDatasets(t *testing.T) {
	cases := []struct {
		name    string
		input   model.TestCase
		wantErr bool
	}{
		{
			name:    "no parameters, no datasets",
			input:   model.TestCase{},
			wantErr: false,
		},
		{
			name: "all datasets have every parameter value",
			input: model.TestCase{
				Parameters: []string{"A", "B"},
				Datasets: []model.Dataset{
					{Name: "case 1", Values: map[string]string{"A": "a1", "B": "b1"}},
				},
			},
			wantErr: false,
		},
		{
			name: "dataset missing one of two parameters",
			input: model.TestCase{
				Parameters: []string{"A", "B"},
				Datasets: []model.Dataset{
					{Name: "case 1", Values: map[string]string{"A": "a1"}},
				},
			},
			wantErr: true,
		},
		{
			name: "second dataset missing a value",
			input: model.TestCase{
				Parameters: []string{"A"},
				Datasets: []model.Dataset{
					{Name: "case 1", Values: map[string]string{"A": "a1"}},
					{Name: "case 2", Values: map[string]string{}},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDatasets(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateDatasets() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
