package objects

import (
	"fmt"
	"strings"
)

func (t *Table) String() string {
	return t.Name
}

func (s *Sequence) String() string {
	return fmt.Sprintf("%s (%s)", s.Name, s.Type)
}

func (c *Column) String() string {
	nullable := "NULL"
	defaulted := ""

	if !c.Nullable {
		nullable = "NOT " + nullable
	}
	if c.Default != "" {
		defaulted = fmt.Sprintf("DEFAULT %s", c.Default)
	}
	if c.MaxLength > 0 {
		return fmt.Sprintf("%s %s(%d) %s %s", c.Name, c.Type, c.MaxLength, nullable, defaulted)
	}
	return fmt.Sprintf("%s %s %s %s", c.Name, c.Type, nullable, defaulted)
}

func (c *Constraint) String() string {
	if c.Type == ContraintTypeForeignKey {
		return fmt.Sprintf("%s %s (%s) REFERENCES %s (%s) ON DELETE %s ON UPDATE %s", c.Name, c.Type, strings.Join(c.Targets, ", "), c.Reference.Table, strings.Join(c.Reference.Columns, ", "), c.OnDelete, c.OnUpdate)
	}
	return fmt.Sprintf("%s %s (%s)", c.Name, c.Type, strings.Join(c.Targets, ", "))
}

func (i *Index) String() string {
	onCols := strings.Join(i.Columns, ", ")
	if i.Unique {
		return fmt.Sprintf("%s (UNIQUE) ON %s", i.Name, onCols)
	}
	return fmt.Sprintf("%s ON %s", i.Name, onCols)
}
