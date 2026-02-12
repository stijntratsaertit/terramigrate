package objects

import (
	"fmt"
	"strings"
)

func (n *Namespace) Valid() error {
	for _, t := range n.Tables {
		err := t.Valid()
		if err != nil {
			return err
		}
	}

	for _, s := range n.Sequences {
		err := s.Valid()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sequence) Valid() error {
	if s.Name == "" {
		return fmt.Errorf("sequence has no name")
	} else if len(s.Name) > 63 {
		return fmt.Errorf("sequence name %s is too long", s.Name)
	} else if s.Type != "bigint" && s.Type != "integer" {
		return fmt.Errorf("sequence type %s is not supported", s.Type)
	}
	return nil
}

func (t *Table) Valid() error {
	if t.Name == "" {
		return fmt.Errorf("table has no name")
	} else if len(t.Name) > 63 {
		return fmt.Errorf("table name %s is too long", t.Name)
	}

	for _, c := range t.Columns {
		err := c.Valid()
		if err != nil {
			return err
		}
	}
	return nil
}

// Returns an error if the column is not valid
func (c *Column) Valid() error {
	if c.Name == "" {
		return fmt.Errorf("column has no name")
	}
	if c.Type == "" {
		return fmt.Errorf("column %s has no type", c.Name)
	}
	if strings.ToUpper(c.Type) == "CHARACTER VARYING" && c.MaxLength == 0 {
		return fmt.Errorf("column %s is of type %s but has no max length", c.Name, c.Type)
	}
	if strings.ToUpper(c.Type) != "CHARACTER VARYING" && c.MaxLength > 0 {
		return fmt.Errorf("column %s is of type %s but has a max length", c.Name, c.Type)
	}
	if !c.Nullable && c.Default == "" && !c.IsPrimaryKey {
		return fmt.Errorf("column %s is not nullable and has no default value", c.Name)
	}
	return nil
}
