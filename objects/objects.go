package objects

import (
	"fmt"
	"strings"
)

type Database struct {
	Name       string       `yaml:"name"`
	Namespaces []*Namespace `yaml:"namespaces"`
}

type Namespace struct {
	Name   string   `yaml:"name"`
	Tables []*Table `yaml:"tables"`
}

type Table struct {
	Name        string        `yaml:"name"`
	Columns     []*Column     `yaml:"columns"`
	Constraints []*Constraint `yaml:"constraints"`
	Indices     []*Index      `yaml:"indices"`
}

func (t *Table) String() string {
	return t.Name
}

type Column struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	MaxLength int    `yaml:"max_length"`
	Nullable  bool   `yaml:"nullable"`
	Default   string `yaml:"default"`
}

func (c *Column) String() string {
	if c.MaxLength > 0 {
		return fmt.Sprintf("%s %s(%d)", c.Name, c.Type, c.MaxLength)
	}
	return fmt.Sprintf("%s %s", c.Name, c.Type)
}

type ConstraintType string

var (
	ContraintTypePrimaryKey ConstraintType = "PRIMARY KEY"
	ContraintTypeUnique     ConstraintType = "UNIQUE"
	ContraintTypeForeignKey ConstraintType = "FOREIGN KEY"
	ContraintTypeCheck      ConstraintType = "CHECK"
	ContraintTypeUnknown    ConstraintType = ""
)

func GetContraintTypeFromCode(code string) ConstraintType {
	switch code {
	case "c":
		return ContraintTypeCheck
	case "f":
		return ContraintTypeForeignKey
	case "p":
		return ContraintTypePrimaryKey
	case "u":
		return ContraintTypeUnique
	default:
		return ContraintTypeUnknown
	}
}

type ConstraintReference struct {
	Table   string
	Columns []string
}

type ContraintAction string

var (
	ContraintActionSetNull    ContraintAction = "SET NULL"
	ContraintActionSetDefault ContraintAction = "SET DEFAULT"
	ContraintActionCascade    ContraintAction = "CASCADE"
	ContraintActionRestrict   ContraintAction = "RESTRICT"
	ContraintActionNoAction   ContraintAction = "NO ACTION"
	ContraintActionUnknown    ContraintAction = ""
)

func GetContraintActionFromCode(code string) ContraintAction {
	switch code {
	case "a":
		return ContraintActionNoAction
	case "d":
		return ContraintActionSetDefault
	case "c":
		return ContraintActionCascade
	case "r":
		return ContraintActionRestrict
	case "n":
		return ContraintActionSetNull
	default:
		return ContraintActionUnknown
	}
}

type Constraint struct {
	Name      string               `yaml:"name"`
	Type      ConstraintType       `yaml:"type"`
	Targets   []string             `yaml:"targets"`
	Reference *ConstraintReference `yaml:"reference"`
	OnDelete  ContraintAction      `yaml:"on_delete"`
	OnUpdate  ContraintAction      `yaml:"on_update"`
}

func (c *Constraint) String() string {
	if c.Type == ContraintTypeForeignKey {
		return fmt.Sprintf("%s %s (%s) REFERENCES %s (%s) ON DELETE %s ON UPDATE %s", c.Name, c.Type, strings.Join(c.Targets, ", "), c.Reference.Table, strings.Join(c.Reference.Columns, ", "), c.OnDelete, c.OnUpdate)
	}
	return fmt.Sprintf("%s %s (%s)", c.Name, c.Type, strings.Join(c.Targets, ", "))
}

type IndexAlgorithm string

var (
	IndexAlgorithmBTree IndexAlgorithm = "btree"
)

type Index struct {
	Name      string         `yaml:"name"`
	Unique    bool           `yaml:"unique"`
	Algorithm IndexAlgorithm `yaml:"algorithm"`
	Columns   []string       `yaml:"columns"`
}

func (i *Index) String() string {
	onCols := strings.Join(i.Columns, ", ")
	if i.Unique {
		return fmt.Sprintf("%s (UNIQUE) ON %s", i.Name, onCols)
	}
	return fmt.Sprintf("%s ON %s", i.Name, onCols)
}
