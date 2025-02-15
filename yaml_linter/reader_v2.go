package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"slices"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/token"
)

/*
	- unique names
	- unique files_names
	- name == file_name
	- ignore .test.yaml files
	- scientists could be duplicated, but project for each scientist should be present only once.
*/

func readDefinitionFilesV2() (map[string]FileContent, error) {
	sets, err := readDefinitionsFromFSV2(DefinitionsFS)
	if err != nil {
		return nil, fmt.Errorf("read featureset definitions from `DefinitionsFS`: %w", err)
	}

	return sets, nil
}

type FileContents map[string]FileContent

type FileContent struct {
	Scientistsset Scientistsset
	Content       []byte
}

func readDefinitionsFromFSV2(defsFS fs.ReadDirFS) (map[string]FileContent, error) {
	files, err := defsFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	fileErrs := make([]FileError, 0)
	scientistsets := make(map[string]FileContent, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := "definitions/" + file.Name()
		content, err := os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		scientistset, err := readFeaturesetYAMLV2(content)
		if err != nil {
			var lintErrs *LintErrors
			if errors.As(err, &lintErrs) {
				fileErrs = append(fileErrs, FileError{
					LintErrors: LintErrors{
						Errors: lintErrs.Errors,
					},
					fileName: fileName,
				})

				continue
			}
			return nil, fmt.Errorf("read file %s: %w", fileName, err)
		}

		scientistsets[file.Name()] = FileContent{
			Scientistsset: scientistset,
			Content:       content,
		}
	}

	if len(fileErrs) > 0 {
		if err := makeAnnotations(fileErrs); err != nil {
			return nil, fmt.Errorf("make annotations: %w", err)
		}
	}

	return scientistsets, nil
}

var (
	EmptyNameErr = errors.New("featureset name must be set")
)

type LintErrors struct {
	Errors []LintError
}

func (e LintErrors) Error() string {
	errs := make([]error, 0, len(e.Errors))
	for _, err := range e.Errors {
		errs = append(errs, err)
	}

	return errors.Join(errs...).Error()
}

type LintError struct {
	Line int
	Err  error
}

func (e LintError) Error() string {
	return string(e.Err.Error())
}

func readFeaturesetYAMLV2(yamlFile []byte) (Scientistsset, error) {
	var yamlMap map[string]any
	if err := yaml.NewDecoder(bytes.NewReader(yamlFile)).Decode(&yamlMap); err != nil {
		return Scientistsset{}, fmt.Errorf("decode yaml: %w", err)
	}

	var defs entitlementsYAML
	if err := mapstructure.Decode(yamlMap, &defs); err != nil {
		return Scientistsset{}, fmt.Errorf("parse decoded yaml: %w", err)
	}

	tokens := lexer.Tokenize(string(yamlFile))
	valErrs := make([]LintError, 0)
	if len(defs.RemainingKeys) > 0 {
		for key := range defs.RemainingKeys {
			line, err := findLineInTokens(key, tokens)
			if err != nil {
				return Scientistsset{}, fmt.Errorf("find line for key %s: %w", key, err)
			}

			valErrs = append(
				valErrs,
				LintError{
					Line: line,
					Err:  fmt.Errorf("unexpected key %s in definitaion", key),
				},
			)
		}
	}

	if defs.Name == "" {
		line, err := findLineInTokens("name", tokens)
		if err != nil {
			// return Scientistsset{}, fmt.Errorf("find line for key 'name': %w", err)
			slog.Error("find line for key 'name'", slog.Any("error", err))
		}
		// if name hasn't been provided --> return Line: 0 (?)

		valErrs = append(valErrs, LintError{Line: line, Err: EmptyNameErr})
	}

	// Validate the file format
	for scientistsName, keys := range defs.Projects {
		if !slices.Contains(defs.Scientists, scientistsName) {
			line, err := findLineInChild(fmt.Sprintf("$.projects.%s", scientistsName), yamlFile)
			if err != nil {
				return Scientistsset{}, fmt.Errorf("find line in child for key %s: %w", scientistsName, err)
			}

			valErrs = append(valErrs, LintError{
				Line: line,
				Err:  fmt.Errorf("scientist %s is not listed in scientists", scientistsName),
			})

			continue
		}

		if len(keys.RemainingKeys) > 0 {
			for key := range keys.RemainingKeys {
				line, err := findLineInChild(fmt.Sprintf("$.projects.%s.%s", scientistsName, key), yamlFile)
				if err != nil {
					return Scientistsset{}, fmt.Errorf("find line in child for key %s: %w", scientistsName, err)
				}

				valErrs = append(valErrs, LintError{
					Line: line,
					Err:  fmt.Errorf("unexpected key %s in definitaion", key),
				})
			}
		}

		if keys.Default <= 0 {
			line, err := findLineInChild(fmt.Sprintf("$.projects.%s.%s", scientistsName, "default"), yamlFile)
			if err != nil {
				return Scientistsset{}, fmt.Errorf("find line for default case in child for %s: %w", scientistsName, err)
			}

			valErrs = append(valErrs, LintError{
				Line: line,
				Err:  fmt.Errorf("default limit for feature %s must be greater than zero", scientistsName),
			})
		}
	}

	fs := Scientistsset{
		Name:        defs.Name,
		Scientitsts: make([]Scientitst, 0, len(defs.Scientists)),
	}

	for _, scientistName := range defs.Scientists {
		scientist := Scientitst{
			Name: scientistName,
		}

		ageCfg, hasAge := defs.Projects[scientistName]
		if hasAge {
			scientist.defaultAge = ageCfg.Default
		}

		fs.Scientitsts = append(fs.Scientitsts, scientist)
	}

	if len(valErrs) > 0 {
		return Scientistsset{}, &LintErrors{Errors: valErrs}
	}

	return fs, nil
}

func findLineInChild(pathToKey string, yamlFile []byte) (int, error) {
	path, err := yaml.PathString(pathToKey)
	if err != nil {
		return 0, err
	}

	node, err := path.ReadNode(bytes.NewReader(yamlFile))
	if err != nil {
		return 0, err
	}

	// we need this small ugly part due to the fact that
	// if `pathToKey` is reffered to the line which is an array or a map
	// the node will start with the line below them
	//. f.e. looking for $.b:
	/*
		b:
		  l: 1
	*/
	// will give us a node `l: 1` and the line (2)
	// so we need to add -1 to the line.
	// it could be avoided with another approach which is currently can't be implemented
	// due to the bug: https://github.com/goccy/go-yaml/issues/659
	addLine := 0
	switch node.Type() {
	case ast.MappingType, ast.SequenceType:
		addLine = -1
	}

	return node.GetToken().Position.Line + addLine, nil

	/* // APPROACH IF THIS BUG FIXED: https://github.com/goccy/go-yaml/issues/659
	sumLine := 0
	yml := string(yamlFile)
	for ind, rootKey := range rootKeys {
		fmt.Printf("YAML: \n%s", yml)

		rootTokens := lexer.Tokenize(yml)
		rootLine, err := findLineInTokens(rootKey, rootTokens)
		if err != nil {
			return 0, fmt.Errorf("looking for root key %s: %w", rootKeys[0], err)
		}
		sumLine += rootLine

		path, err := yaml.PathString(fmt.Sprintf("$.%s", rootKey))
		if err != nil {
			return 0, fmt.Errorf("finding path: %w", err)
		}

		node, err := path.ReadNode(strings.NewReader(yml))
		if err != nil {
			return 0, fmt.Errorf("reading node: %w", err)
		}

		yml = node.String()
		if ind != len(rootKeys)-1 {
			continue
		}

		tokens := lexer.Tokenize(node.String())
		line, err := findLineInTokens(key, tokens)
		if err != nil {
			return 0, fmt.Errorf("lookign for key %s: %w", key, err)
		}
		sumLine += line
	}

	return sumLine, nil
	*/
}

func findLineInTokens(key string, tokens token.Tokens) (int, error) {
	for _, token := range tokens {
		if token.Value == key {
			return token.Position.Line, nil
		}
	}

	return 0, errors.New("key not found")
}

func findLine(pathToKey string, yamlContent []byte) (int, error) {
	path, err := yaml.PathString(pathToKey)
	if err != nil {
		return 0, err
	}

	ast, err := path.ReadNode(bytes.NewReader(yamlContent))
	if err != nil {
		return 0, err
	}

	return ast.GetToken().Position.Line, nil
}
