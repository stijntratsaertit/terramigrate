package state

import (
	"fmt"
	"os"
	"stijntratsaertit/terramigrate/objects"

	"gopkg.in/yaml.v2"
)

type request struct {
	Namespaces []*objects.Namespace `yaml:"namespaces"`
}

func LoadYAML(path string) (*request, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", path)
	}

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %v", path, err)
	}

	req := &request{}
	err = yaml.Unmarshal(yamlFile, req)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml: %v", err)
	}

	return req, nil
}

func (s *State) ExportYAML(path string) error {
	yamlFile, err := yaml.Marshal(request{Namespaces: s.Database.Namespaces})
	if err != nil {
		return fmt.Errorf("could not marshal yaml: %v", err)
	}

	err = os.WriteFile(path, yamlFile, 0644)
	if err != nil {
		return fmt.Errorf("could not write file %s: %v", path, err)
	}

	return nil
}
