package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errUniqueName              = errors.New("the name should be unuque")
	errFileNameMatch           = errors.New("the name should be the same as the file name")
	errOnlyOneProjectIsAllowed = errors.New("only one default project is allowed")
)

type FileErrors []FileError

func (errs FileErrors) Error() string {
	return "KEK"
}

func (fcs *FileContents) Validate() error {
	fileErrs := make([]FileError, 0)
	uniqueNameErrs, err := fcs.uniqueNames()
	if err != nil {
		return fmt.Errorf("validating unique names: %w", err)
	}
	fileErrs = append(fileErrs, uniqueNameErrs...)

	matchErrs, err := fcs.fileNameMatch()
	if err != nil {
		return fmt.Errorf("validating names on matching with file name: %w", err)
	}
	fileErrs = append(fileErrs, matchErrs...)

	onePrjectErrs, err := fcs.projectOnlyOnce()
	if err != nil {
		return fmt.Errorf("validating only one project is allowed: %w", err)
	}
	fileErrs = append(fileErrs, onePrjectErrs...)

	if len(fileErrs) > 0 {
		ferrs := FileErrors(fileErrs)
		return &ferrs
	}

	return nil
}

type uniqueName struct {
	fileName string
	marked   bool
}

func (fcs *FileContents) uniqueNames() ([]FileError, error) {
	fileErrs := make([]FileError, 0, len(*fcs))
	uniqueNames := make(map[string]uniqueName, len(*fcs))
	for fileName, fileContent := range *fcs {
		if existingUniqeuName, ok := uniqueNames[fileContent.Scientistsset.Name]; ok {
			lintErr, err := makeUniqueNameLintErr("$.name", fileContent.Content)
			if err != nil {
				return nil, err
			}

			fileErrs = append(fileErrs, FileError{
				LintErrors: LintErrors{
					Errors: []LintError{lintErr},
				},
				fileName: fileName,
			})

			if !existingUniqeuName.marked {
				m := *fcs
				cnt, ok := m[existingUniqeuName.fileName]
				if !ok {
					return nil, fmt.Errorf("file %s not found in the internal map", existingUniqeuName.fileName)
				}

				lintErr, err := makeUniqueNameLintErr("$.name", cnt.Content)
				if err != nil {
					return nil, err
				}

				fileErrs = append(fileErrs, FileError{
					LintErrors: LintErrors{
						Errors: []LintError{lintErr},
					},
					fileName: existingUniqeuName.fileName,
				})

				existingUniqeuName.marked = true
				uniqueNames[fileContent.Scientistsset.Name] = existingUniqeuName
			}

			continue
		}

		uniqueNames[fileContent.Scientistsset.Name] = uniqueName{
			fileName: fileName,
			marked:   false,
		}
	}

	return fileErrs, nil
}

func (fcs *FileContents) fileNameMatch() ([]FileError, error) {
	fileErrs := make([]FileError, 0, len(*fcs))
	for fileName, fileContent := range *fcs {
		if strings.HasPrefix(fileName, fileContent.Scientistsset.Name) {
			continue
		}

		lintErr, err := makeFileNameMatchLintErr("$.name", fileContent.Content)
		if err != nil {
			return nil, err
		}

		fileErrs = append(fileErrs, FileError{
			LintErrors: LintErrors{
				Errors: []LintError{lintErr},
			},
			fileName: fileName,
		})
	}

	return fileErrs, nil
}

type onlyOneProject struct {
	fileName string
	marked   bool
}

func (fcs *FileContents) projectOnlyOnce() ([]FileError, error) {
	fileErrs := make([]FileError, 0, len(*fcs))
	uniqueByProjects := make(map[string]onlyOneProject, len(*fcs))
	for fileName, fileContent := range *fcs {
		for _, scientist := range fileContent.Scientistsset.Scientitsts {
			if scientist.defaultProjects == nil {
				continue
			}

			if existing, ok := uniqueByProjects[scientist.Name]; ok {
				key := fmt.Sprintf("$.projects.%s", scientist.Name)
				lintErr, err := makeOnlyOneProjectIsAllowedLintErr(key, fileContent.Content)
				if err != nil {
					return nil, err
				}

				fileErrs = append(fileErrs, FileError{
					LintErrors: LintErrors{
						Errors: []LintError{lintErr},
					},
					fileName: fileName,
				})

				if !existing.marked {
					m := *fcs
					cnt, ok := m[existing.fileName]
					if !ok {
						return nil, fmt.Errorf("file %s not found in the internal map", existing.fileName)
					}

					key := fmt.Sprintf("$.projects.%s", scientist.Name)
					lintErr, err := makeOnlyOneProjectIsAllowedLintErr(key, cnt.Content)
					if err != nil {
						return nil, err
					}

					fileErrs = append(fileErrs, FileError{
						LintErrors: LintErrors{
							Errors: []LintError{lintErr},
						},
						fileName: existing.fileName,
					})

					existing.marked = true
					uniqueByProjects[scientist.Name] = existing
				}

				continue
			}

			uniqueByProjects[scientist.Name] = onlyOneProject{
				fileName: fileName,
				marked:   false,
			}
		}
	}

	return fileErrs, nil
}

func makeUniqueNameLintErr(pathToKey string, content []byte) (LintError, error) {
	line, err := findLine(pathToKey, content)
	if err != nil {
		return LintError{}, fmt.Errorf("finding the key in the content: %w", err)
	}

	return LintError{
		Line: line,
		Err:  errUniqueName,
	}, nil
}

func makeFileNameMatchLintErr(pathToKey string, content []byte) (LintError, error) {
	line, err := findLine(pathToKey, content)
	if err != nil {
		return LintError{}, fmt.Errorf("finding the key in the content: %w", err)
	}

	return LintError{
		Line: line,
		Err:  errFileNameMatch,
	}, nil
}

func makeOnlyOneProjectIsAllowedLintErr(pathToKey string, content []byte) (LintError, error) {
	line, err := findLine(pathToKey, content)
	if err != nil {
		return LintError{}, fmt.Errorf("finding the key in the content: %w", err)
	}

	return LintError{
		Line: line,
		Err:  errOnlyOneProjectIsAllowed,
	}, nil
}

/*
	- ignore .test.yaml files
	- scientists could be duplicated, but project for each scientist should be present only once.
	- mark as a warning removed featuresets/features
*/
