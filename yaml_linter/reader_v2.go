package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func f() {
	file, err := os.ReadFile("./definitions/test.yaml")
	if err != nil {
		panic(err)
	}
	yml := string(file)

	var v struct {
		A []int
		B string
	}
	if err := yaml.Unmarshal(file, &v); err != nil {
		panic(err)
	}

	if v.A[0] != 2 {
		// output error with YAML source
		path, err := yaml.PathString("$.a[0]")
		if err != nil {
			panic(err)
		}

		source, err := path.AnnotateSource([]byte(yml), true)
		if err != nil {
			panic(err)
		}
		fmt.Printf("a value expected 2 but actual %d:\n%s\n", v.A, string(source))
	}

	/*
		tokens := lexer.Tokenize(yml)
		for _, token := range tokens {
			fmt.Println(token)
		}
	*/

	ann := annotation{
		File:            "./yaml_linter/definitions/test.yaml",
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
