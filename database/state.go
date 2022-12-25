package database

import (
	"fmt"
)

type state struct {
	Tables []*Table
}

func (s *state) String() string {
	result := "\n"
	for _, table := range s.Tables {
		result += fmt.Sprintf("%v\n", table.String())
		for _, column := range table.Columns {
			result += fmt.Sprintf("\t%v\n", column.String())
		}
		for _, constraint := range table.Contraints {
			result += fmt.Sprintf("\t%v\n", constraint.String())
		}
	}
	return result
}
