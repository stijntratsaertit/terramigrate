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
	if c.Type == ConstraintTypeForeignKey {
		return fmt.Sprintf("%s %s (%s) REFERENCES %s (%s) ON DELETE %s ON UPDATE %s", c.Name, c.Type, strings.Join(c.Targets, ", "), c.Reference.Table, strings.Join(c.Reference.Columns, ", "), c.OnDelete, c.OnUpdate)
	}
	return fmt.Sprintf("%s %s (%s)", c.Name, c.Type, strings.Join(c.Targets, ", "))
}

func (c *Constraint) SQL() string {
	base := fmt.Sprintf("CONSTRAINT %s %s (%s)", c.Name, c.Type, strings.Join(c.Targets, ", "))
	if c.Type == ConstraintTypeForeignKey && c.Reference != nil {
		base += fmt.Sprintf(" REFERENCES %s (%s)", c.Reference.Table, strings.Join(c.Reference.Columns, ", "))
		if c.OnDelete != "" && c.OnDelete != ConstraintActionUnknown {
			base += fmt.Sprintf(" ON DELETE %s", c.OnDelete)
		}
		if c.OnUpdate != "" && c.OnUpdate != ConstraintActionUnknown {
			base += fmt.Sprintf(" ON UPDATE %s", c.OnUpdate)
		}
	}
	return base
}

func (c *Constraint) Equal(other *Constraint) bool {
	if c.Type != other.Type {
		return false
	}
	if len(c.Targets) != len(other.Targets) {
		return false
	}
	for i, t := range c.Targets {
		if t != other.Targets[i] {
			return false
		}
	}
	if c.Type == ConstraintTypeForeignKey {
		if c.Reference == nil || other.Reference == nil {
			return c.Reference == other.Reference
		}
		if c.Reference.Table != other.Reference.Table {
			return false
		}
		if len(c.Reference.Columns) != len(other.Reference.Columns) {
			return false
		}
		for i, col := range c.Reference.Columns {
			if col != other.Reference.Columns[i] {
				return false
			}
		}
		if c.OnDelete != other.OnDelete || c.OnUpdate != other.OnUpdate {
			return false
		}
	}
	return true
}

func (i *Index) String() string {
	onCols := strings.Join(i.Columns, ", ")
	if i.Unique {
		return fmt.Sprintf("%s (UNIQUE) ON %s", i.Name, onCols)
	}
	return fmt.Sprintf("%s ON %s", i.Name, onCols)
}

func (i *Index) Equal(other *Index) bool {
	if i.Unique != other.Unique {
		return false
	}
	if i.Algorithm != other.Algorithm {
		return false
	}
	if len(i.Columns) != len(other.Columns) {
		return false
	}
	for idx, col := range i.Columns {
		if col != other.Columns[idx] {
			return false
		}
	}
	return true
}
