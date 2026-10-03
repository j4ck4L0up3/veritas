package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"sigs.k8s.io/yaml"
)

//go:embed config.schema.json
var schemaJSON []byte

func compileSchema() (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat() // otherwise "format": "ipv4" is annotation-only
	if err := c.AddResource("config.schema.json", doc); err != nil {
		return nil, err
	}
	return c.Compile("config.schema.json")
}

func validate(sch *jsonschema.Schema, yamlData []byte) error {
	jsonData, err := yaml.YAMLToJSON(yamlData)
	if err != nil {
		return fmt.Errorf("yaml: %w", err)
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	return sch.Validate(inst)
}
