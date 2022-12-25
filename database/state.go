package database

import (
	"fmt"
)

type state struct {
	Tables []*Table
}

func (s *state) String() string {
	result := "tables:\n"
	for _, table := range s.Tables {
		result += fmt.Sprintf("- name: %v\n", table.String())
		if len(table.Columns) != 0 {
			result += "  columns:\n"
			for _, column := range table.Columns {
				result += fmt.Sprintf("  - %v\n", column.String())
			}
		}
		if len(table.Constraints) != 0 {
			result += "  constraints:\n"
			for _, constraint := range table.Constraints {
				result += fmt.Sprintf("  - %v\n", constraint.String())
			}
		}
		if len(table.Indices) != 0 {
			result += "  indices:\n"
			for _, index := range table.Indices {
				result += fmt.Sprintf("  - %v\n", index.String())
			}
		}
	}
	return result
}
