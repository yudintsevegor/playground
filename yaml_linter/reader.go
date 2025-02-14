package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"yudintsevegor/playground/yaml_linter/definitions"

	"github.com/go-viper/mapstructure/v2"
	"gopkg.in/yaml.v3"
)

type entitlementsYAML struct {
	Type       string   `mapstructure:"type"`
	Name       string   `mapstructure:"name"`
	Scientists []string `mapstructure:"scientists"`
	Projects   map[string]struct {
		Default       int            `mapstructure:"default"`
		RemainingKeys map[string]any `mapstructure:",remain"`
	} `mapstructure:"projects"`

	RemainingKeys map[string]any `mapstructure:",remain"`
}

type Scientistsset struct {
	Name        string
	Scientitsts []Scientitst
}

type Scientitst struct {
	Name       string
	defaultAge int
}

var DefinitionsFS fs.ReadDirFS = definitions.EmbedFS

func readDefinitionFiles() (map[string]Scientistsset, error) {
	sets, err := readDefinitionsFromFS(DefinitionsFS)
	if err != nil {
		return nil, fmt.Errorf("read featureset definitions from `DefinitionsFS`: %w", err)
	}

	return sets, nil
}

func readDefinitionsFromFS(defsFS fs.ReadDirFS) (map[string]Scientistsset, error) {
	files, err := defsFS.ReadDir(".")
	if err != nil {
		return nil, err
	}

	scientistsets := make(map[string]Scientistsset, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		f, err := defsFS.Open(file.Name())
		if err != nil {
			return nil, fmt.Errorf("open file: %w", err)
		}

		scientistset, err := readFeaturesetYAML(f)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", file.Name(), err)
		}

		scientistsets[file.Name()] = scientistset
	}

	return scientistsets, nil
}

func readFeaturesetYAML(r io.Reader) (Scientistsset, error) {
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	var yamlMap map[string]any
	if err := yaml.NewDecoder(tee).Decode(&yamlMap); err != nil {
		return Scientistsset{}, fmt.Errorf("decode yaml: %w", err)
	}

	var node yaml.Node
	if err := yaml.NewDecoder(&buf).Decode(&node); err != nil {
		return Scientistsset{}, fmt.Errorf("decode yaml to node: %w", err)
	}
	readNodes(node)

	var defs entitlementsYAML
	if err := mapstructure.Decode(yamlMap, &defs); err != nil {
		return Scientistsset{}, fmt.Errorf("parse decoded yaml: %w", err)
	}

	if len(defs.RemainingKeys) > 0 {
		keys := slices.Collect(maps.Keys(defs.RemainingKeys))
		return Scientistsset{}, fmt.Errorf("unexpected keys [%s] in entitlement definition", strings.Join(keys, ","))
	}

	if defs.Name == "" {
		return Scientistsset{}, errors.New("featureset name must be set")
	}

	// Validate the file format
	for lf, lc := range defs.Projects {
		if !slices.Contains(defs.Scientists, lf) {
			return Scientistsset{}, fmt.Errorf("limit defined for undefined feature %s", lf)
		}

		if len(lc.RemainingKeys) > 0 {
			keys := slices.Collect(maps.Keys(lc.RemainingKeys))
			return Scientistsset{}, fmt.Errorf("unexpected keys [%s] in limits definition", strings.Join(keys, ","))
		}

		if lc.Default <= 0 {
			return Scientistsset{}, fmt.Errorf("default limit for feature %s must be greater than zero", lf)
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

	return fs, nil
}

func readNodes(node yaml.Node) {
	for _, n := range node.Content {
		slog.Info(
			"Node",
			slog.Any("value", n.Value),
			slog.Any("line", n.Line),
			slog.Any("node", n),
		)
		readNodes(*n)
	}
}
