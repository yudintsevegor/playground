package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
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

type FileContent struct {
	Scientistsset Scientistsset
	Content       []byte
}

func readDefinitionsFromFSV2(defsFS fs.ReadDirFS) (map[string]FileContent, error) {
	files, err := defsFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	scientistsets := make(map[string]FileContent, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := os.ReadFile(file.Name())
		if err != nil {
			return nil, err
		}

		scientistset, err := readFeaturesetYAMLV2(content)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", file.Name(), err)
		}

		scientistsets[file.Name()] = FileContent{
			Scientistsset: scientistset,
			Content:       content,
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
			return Scientistsset{}, fmt.Errorf("find line for key 'name': %w", err)
		}
		// if name hasn't been provided --> return Line: 0 (?)

		valErrs = append(valErrs, LintError{Line: line, Err: EmptyNameErr})
	}

	// Validate the file format
	for scientistsName, keys := range defs.Projects {
		if !slices.Contains(defs.Scientists, scientistsName) {
			line, err := findLineInChild(scientistsName, []string{"projects"}, yamlFile)
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
				line, err := findLineInChild(key, []string{"projects"}, yamlFile)
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
			line, err := findLineInChild("default", []string{"projects", scientistsName}, yamlFile)
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

func findLineInChild(key string, rootKeys []string, yamlFile []byte) (int, error) {
	rootTokens := lexer.Tokenize(string(yamlFile))
	rootLine, err := findLineInTokens(rootKeys[0], rootTokens)
	if err != nil {
		return 0, fmt.Errorf("looking for root key %s: %w", rootKeys[0], err)
	}

	path, err := yaml.PathString(fmt.Sprintf("$.%s", strings.Join(rootKeys, ".")))
	if err != nil {
		return 0, err
	}

	node, err := path.ReadNode(bytes.NewReader(yamlFile))
	if err != nil {
		return 0, err
	}

	tokens := lexer.Tokenize(node.String())

	line, err := findLineInTokens(key, tokens)
	if err != nil {
		return 0, fmt.Errorf("lookign for key %s: %w", key, err)
	}

	return rootLine + line + len(rootKeys) - 1, nil
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

	log.Printf("ast: %+v", ast.GetToken())

	return ast.GetToken().Position.Line, nil
}

func makeAnnotation() {
	ann := annotation{
		File:            "yaml_linter/definitions/test.yaml",
		Line:            2,
		Title:           "Test Annotation",
		Message:         "It's a test annotation",
		AnnotationLevel: "failure", // Can be one of: notice, warning, failure
	}

	anns := []annotation{ann}
	body, err := json.Marshal(anns)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = os.WriteFile("./annotations.json", body, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
}

type annotation struct {
	File            string `json:"file"`
	Line            int    `json:"line"`
	Title           string `json:"title"`
	Message         string `json:"message"`
	AnnotationLevel string `json:"annotation_level"`
}

/*
[
  {
    file: "path/to/file.js",
    line: 5,
    title: "title for my annotation",
    message: "my message",
    annotation_level: "failure"
  }
]
*/
