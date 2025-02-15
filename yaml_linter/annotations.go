package main

import (
	"encoding/json"
	"os"
)

type FileError struct {
	LintErrors
	fileName string
}

type annotation struct {
	File            string `json:"file"`
	Line            int    `json:"line"`
	Title           string `json:"title"`
	Message         string `json:"message"`
	AnnotationLevel string `json:"annotation_level"`
}

func makeAnnotations(fileErrs []FileError) error {
	anns := make([]annotation, 0, len(fileErrs))
	for _, fileErr := range fileErrs {
		for _, lintErr := range fileErr.Errors {
			anns = append(anns, annotation{
				File:            fileErr.fileName,
				Line:            lintErr.Line,
				Title:           "Lint Error",
				Message:         lintErr.Error(),
				AnnotationLevel: "failure",
			})
		}
	}

	body, err := json.Marshal(anns)
	if err != nil {
		return err
	}

	err = os.WriteFile("./annotations.json", body, 0644)
	if err != nil {
		return err
	}

	return nil
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
