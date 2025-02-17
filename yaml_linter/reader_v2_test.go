package main

import (
	"fmt"
	"testing"

	"errors"

	"github.com/stretchr/testify/require"
)

const (
	unknownKeysContent = `name: kek

cupakabra: 123
scientists:
 - feynaman
 - einstein
cheburashka:
 - 1
 - 2
#comment: kek
`
	emptyNameContent = `name:

scientists:
 - feynaman
 - einstein
#comment: kek
`

	unknownKeysEmptyNameContent = `name: 

cupakabra: 123
scientists:
 - feynaman
 - einstein
cheburashka:
 - 1
 - 2
#comment: kek
`

	unknownInProjectsContent = `name: kek 

scientists:
 - feynaman
 - einstein
#comment: kek
projects:
  newton:
    default: 1
`
	unknownKeyInProjectsContent = `name: kek 

scientists:
 - feynman
 - einstein
#comment: kek
projects:
  feynman:
    unknown: 1
    default: 2
`

	invalidDefaultContent = `name: kek 

scientists:
 - feynman
 - einstein
#comment: kek
projects:
  #maxwell:
    #someField: 1
    #default: 0
  einstein:
    default: 11
  feynman:
    #someField: 1
    default: 0
`
)

func Test_readFeaturesetYAMLV2(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		lintErrs []LintError
	}{
		{
			name:    "unknown keys",
			content: unknownKeysContent,
			wantErr: true,
			lintErrs: []LintError{
				{
					Line: 3,
					Err:  fmt.Errorf("unexpected key %s in definitaion", "cupakabra"),
				},
				{
					Line: 7,
					Err:  fmt.Errorf("unexpected key %s in definitaion", "cheburashka"),
				},
			},
		},
		{
			name:    "empty name",
			content: emptyNameContent,
			wantErr: true,
			lintErrs: []LintError{
				{
					Line: 1,
					Err:  EmptyNameErr,
				},
			},
		},
		{
			name:    "empty name and unknown keys",
			content: unknownKeysEmptyNameContent,
			wantErr: true,
			lintErrs: []LintError{
				{
					Line: 1,
					Err:  EmptyNameErr,
				},
				{
					Line: 3,
					Err:  fmt.Errorf("unexpected key %s in definitaion", "cupakabra"),
				},
				{
					Line: 7,
					Err:  fmt.Errorf("unexpected key %s in definitaion", "cheburashka"),
				},
			},
		},
		{
			name:    "unknown project",
			content: unknownInProjectsContent,
			wantErr: true,
			lintErrs: []LintError{
				{
					Line: 8,
					Err:  fmt.Errorf("scientist %s is not listed in scientists", "newton"),
				},
			},
		},
		{
			name:    "unknown key in projects",
			content: unknownKeyInProjectsContent,
			wantErr: true,
			lintErrs: []LintError{
				{
					Line: 9,
					Err:  fmt.Errorf("unexpected key %s in definitaion", "unknown"),
				},
			},
		},
		{
			name:    "invalid default",
			content: invalidDefaultContent,
			wantErr: true,
			lintErrs: []LintError{
				{
					Line: 15,
					Err:  fmt.Errorf("default limit for feature %s must be greater than zero", "feynman"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readFeaturesetYAMLV2([]byte(tt.content))
			if tt.wantErr {
				require.Error(t, err)
				if len(tt.lintErrs) > 0 {
					errs := new(LintErrors)
					errors.As(err, &errs)
					require.ElementsMatch(t, tt.lintErrs, errs.Errors, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
