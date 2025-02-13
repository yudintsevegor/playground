package main

import (
	"fmt"
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
}
