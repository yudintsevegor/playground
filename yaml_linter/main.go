package main

import (
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/lexer"
)

func main() {
	f()
	return
	scientists, err := readDefinitionFiles()
	if err != nil {
		slog.Error("error reading definition files", slog.Any("error", err))
	}

	slog.Info("Scientists", slog.Any("scientists", scientists))
}

func f() {
	file, err := os.ReadFile("./definitions/test.yaml")
	if err != nil {
		panic(err)
	}
	yml := string(file)

	var v struct {
		A []int  `yaml:"a"`
		B string `yaml:"b"`
	}
	if err := yaml.Unmarshal(file, &v); err != nil {
		panic(err)
	}

	log.Println(v.A)
	if v.A[2] != 2 {
		// output error with YAML source
		path, err := yaml.PathString("$.a[2]")
		if err != nil {
			panic(err)
		}

		ast, err := path.ReadNode(bytes.NewReader(file))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("ast: %+v", ast.GetToken().Position.Line)

		source, err := path.AnnotateSource([]byte(yml), true)
		if err != nil {
			panic(err)
		}
		fmt.Printf("a value expected 2 but actual %d:\n%s\n", v.A, string(source))
	}

	/**
	tokens := lexer.Tokenize(yml)
	for _, token := range tokens {
		fmt.Printf("%+v\n\n", token)
		fmt.Printf("VALUE: %v\n\n", token.Value)
		fmt.Printf("LINE: %v\n\n", token.Position.Line)
	}
	/**/

	path, err := yaml.PathString("$.c")
	if err != nil {
		panic(err)
	}
	node, err := path.ReadNode(bytes.NewReader(file))
	if err != nil {
		panic(err)
	}

	tokens := lexer.Tokenize(node.String())
	for _, token := range tokens {
		fmt.Printf("%+v\n\n", token)
		// fmt.Printf("VALUE: %v\n\n", token.Value)
		fmt.Printf("LINE: %v\n\n", token.Position.Line)
	}

	if false {
		makeAnnotation()
	}
}
