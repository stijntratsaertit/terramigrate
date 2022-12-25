package database

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	indexDefinitionRegex = regexp.MustCompile(`CREATE( UNIQUE)? INDEX (\w+) ON (\w+)\.(\w+) USING (\w+) \((.+)\)`)
)

func parseIndexDefinition(indexDef string) (*Index, error) {
	matches := indexDefinitionRegex.FindStringSubmatch(indexDef)
	if len(matches) != 7 {
		return nil, fmt.Errorf("could not extract index definition from %s", indexDef)
	}

	return &Index{
		Name:      matches[2],
		Unique:    matches[1] != "",
		Algorithm: IndexAlgorithm(matches[5]),
		Columns:   strings.Split(matches[6], ", "),
	}, nil
}
