package main

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-yaml"
)

var content = `
name: test
a:
  - 0
  - 1
b:
  #l: 1
  d: 1
`

type yamlContent struct {
	Name string         `yaml:"name"`
	A    []int          `yaml:"a"`
	B    map[string]int `yaml:"b"`
}

func example() {
	var y yamlContent
	err := yaml.Unmarshal([]byte(content), &y)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", y) // {Name:test A:[0 1] B:map[d:1]} - expected.

	path, err := yaml.PathString("$.b")
	if err != nil {
		panic(err)
	}
	fmt.Println("PATH", path.String()) // d: 1 - not expected.
	node, err := path.ReadNode(bytes.NewReader([]byte(content)))
	if err != nil {
		panic(err)
	}
	fmt.Println("NODE", node.String()) // d: 1 - not expected.
	/*
		#l: 1
		d: 1
	*/

}
