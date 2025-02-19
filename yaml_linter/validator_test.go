package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testOneContent = `name: test_1

scientists:
 - newton
 - einstein	

projects:
 newton:
   default: 1   
`
	testTwoContent = `type: sets

name: test_1

scientists:
 - newton
 - feynman

projects:
 newton:
   default: 1
`
	testThreeContent = `type: sets

name: test_3

scientists:
 - curie 
 - hawking
`
)

func Test_uniqueNames(t *testing.T) {
	tests := []struct {
		name     string
		fcs      FileContents
		expected []FileError
	}{
		{
			name: "happy path",
			fcs: map[string]FileContent{
				"test_1.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_1",
						Scientitsts: nil, // not important for the tests
					},
				},
				"test_2.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_2",
						Scientitsts: nil, // not important for the tests
					},
				},
				"test_3.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_3",
						Scientitsts: nil, // not important for the tests
					},
				},
			},
			expected: nil,
		},
		{
			name: "not unique names",
			fcs: map[string]FileContent{
				"test_1.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_1",
						Scientitsts: nil, // not important for the tests
					},
					Content: []byte(testOneContent),
				},
				"test_2.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_1",
						Scientitsts: nil, // not important for the tests
					},
					Content: []byte(testTwoContent),
				},
				"test_3.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_3",
						Scientitsts: nil, // not important for the tests
					},
					Content: []byte(testThreeContent),
				},
			},
			expected: []FileError{
				{
					LintErrors: LintErrors{
						Errors: []LintError{
							{
								Line: 1,
								Err:  errUniqueName,
							},
						},
					},
					fileName: "test_1.yaml",
				},
				{
					LintErrors: LintErrors{
						Errors: []LintError{
							{
								Line: 3,
								Err:  errUniqueName,
							},
						},
					},
					fileName: "test_2.yaml",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fcs.uniqueNames()
			require.NoError(t, err)

			/*
				require.Equal(t, len(tt.expected), len(result))
				for ind := 0; ind <= len(tt.expected)-1; ind++ {
					sort.Slice(tt.expected[ind].LintErrors.Errors, func(i, j int) bool {
						return tt.expected[ind].LintErrors.Errors[i].Line <
							tt.expected[ind].LintErrors.Errors[j].Line
					})
					sort.Slice(result[ind].LintErrors.Errors, func(i, j int) bool {
						return result[ind].LintErrors.Errors[i].Line <
							result[ind].LintErrors.Errors[j].Line
					})
				}
			*/

			require.ElementsMatch(t, tt.expected, result)
		})
	}
}

func Test_fileNameMatch(t *testing.T) {
	tests := []struct {
		name     string
		fcs      FileContents
		expected []FileError
	}{
		{
			name: "happy path",
			fcs: map[string]FileContent{
				"test_1.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_1",
						Scientitsts: nil, // not important for the tests
					},
				},
				"test_2.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_2",
						Scientitsts: nil, // not important for the tests
					},
				},
				"test_3.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_3",
						Scientitsts: nil, // not important for the tests
					},
				},
			},
			expected: nil,
		},
		{
			name: "not matching names",
			fcs: map[string]FileContent{
				"test_1.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_1",
						Scientitsts: nil, // not important for the tests
					},
					Content: []byte(testOneContent),
				},
				"test_2.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_1",
						Scientitsts: nil, // not important for the tests
					},
					Content: []byte(testTwoContent),
				},
				"test_3.yaml": {
					Scientistsset: Scientistsset{
						Name:        "test_3",
						Scientitsts: nil, // not important for the tests
					},
					Content: []byte(testThreeContent),
				},
			},
			expected: []FileError{
				{
					LintErrors: LintErrors{
						Errors: []LintError{
							{
								Line: 3,
								Err:  errFileNameMatch,
							},
						},
					},
					fileName: "test_2.yaml",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fcs.fileNameMatch()
			require.NoError(t, err)

			/*
				require.Equal(t, len(tt.expected), len(result))
				for ind := 0; ind <= len(tt.expected)-1; ind++ {
					sort.Slice(tt.expected[ind].LintErrors.Errors, func(i, j int) bool {
						return tt.expected[ind].LintErrors.Errors[i].Line <
							tt.expected[ind].LintErrors.Errors[j].Line
					})
					sort.Slice(result[ind].LintErrors.Errors, func(i, j int) bool {
						return result[ind].LintErrors.Errors[i].Line <
							result[ind].LintErrors.Errors[j].Line
					})
				}
			*/

			require.ElementsMatch(t, tt.expected, result)
		})
	}
}

func Test_projectOnlyOnce(t *testing.T) {
	tests := []struct {
		name     string
		fcs      FileContents
		expected []FileError
	}{
		{
			name: "happy path",
			fcs: map[string]FileContent{
				"test_1.yaml": {
					Scientistsset: Scientistsset{
						Name: "test_1",
						Scientitsts: []Scientitst{
							{
								Name: "newton",
							},
							{
								Name: "einstein",
							},
						},
					},
				},
				"test_2.yaml": {
					Scientistsset: Scientistsset{
						Name: "test_2",
						Scientitsts: []Scientitst{
							{
								Name: "feynman",
							},
							{
								Name: "newton",
							},
						},
					},
				},
				"test_3.yaml": {
					Scientistsset: Scientistsset{
						Name: "test_3",
						Scientitsts: []Scientitst{
							{
								Name: "curie",
							},
							{
								Name: "hawking",
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "fails when 2 default values",
			fcs: map[string]FileContent{
				"test_1.yaml": {
					Scientistsset: Scientistsset{
						Name: "test_1",
						Scientitsts: []Scientitst{
							{
								Name:            "newton",
								defaultProjects: refInt(1),
							},
							{
								Name: "einstein",
							},
						},
					},
					Content: []byte(testOneContent),
				},
				"test_2.yaml": {
					Scientistsset: Scientistsset{
						Name: "test_2",
						Scientitsts: []Scientitst{
							{
								Name: "feynman",
							},
							{
								Name:            "newton",
								defaultProjects: refInt(1),
							},
						},
					},
					Content: []byte(testTwoContent),
				},
				"test_3.yaml": {
					Scientistsset: Scientistsset{
						Name: "test_3",
						Scientitsts: []Scientitst{
							{
								Name: "curie",
							},
							{
								Name: "hawking",
							},
						},
					},
				},
			},
			expected: []FileError{
				{
					LintErrors: LintErrors{
						Errors: []LintError{
							{
								Line: 8,
								Err:  errOnlyOneProjectIsAllowed,
							},
						},
					},
					fileName: "test_1.yaml",
				},
				{
					LintErrors: LintErrors{
						Errors: []LintError{
							{
								Line: 10,
								Err:  errOnlyOneProjectIsAllowed,
							},
						},
					},
					fileName: "test_2.yaml",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fcs.projectOnlyOnce()
			require.NoError(t, err)

			/*
				require.Equal(t, len(tt.expected), len(result))
				for ind := 0; ind <= len(tt.expected)-1; ind++ {
					sort.Slice(tt.expected[ind].LintErrors.Errors, func(i, j int) bool {
						return tt.expected[ind].LintErrors.Errors[i].Line <
							tt.expected[ind].LintErrors.Errors[j].Line
					})
					sort.Slice(result[ind].LintErrors.Errors, func(i, j int) bool {
						return result[ind].LintErrors.Errors[i].Line <
							result[ind].LintErrors.Errors[j].Line
					})
				}
			*/

			require.ElementsMatch(t, tt.expected, result)
		})
	}
}

func refInt(in int) *int {
	return &in
}
