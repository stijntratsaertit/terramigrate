package parser

import (
	"fmt"
	"os"
	"stijntratsaertit/terramigrate/objects"

	"gopkg.in/yaml.v2"
)

type Request struct {
	Namespaces []*objects.Namespace `yaml:"namespaces"`
}

func LoadYAML(path string) (*Request, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", path)
	}

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %v", path, err)
	}

	request := &Request{}
	err = yaml.Unmarshal(yamlFile, request)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml: %v", err)
	}

	return request, nil
}

func ExportYAML(path string, request *Request) error {
	yamlFile, err := yaml.Marshal(request)
	if err != nil {
		return fmt.Errorf("could not marshal yaml: %v", err)
	}

	err = os.WriteFile(path, yamlFile, 0644)
	if err != nil {
		return fmt.Errorf("could not write file %s: %v", path, err)
	}

	return nil
}
