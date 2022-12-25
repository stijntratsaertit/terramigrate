package database

import (
	"fmt"
)

type DatabaseObject interface {
	String() string
}

type Table struct {
	Name       string
	Columns    []*Column
	Contraints []*Constraint
	Indices    []*Index
}

func (t *Table) String() string {
	return t.Name
}

type Column struct {
	Name      string
	Type      string
	MaxLength int
	Nullable  bool
	Default   string
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
)

type ConstraintReference struct {
	Table  string
	Column string
}

type ContraintAction string

var (
	ContraintActionSetNull    ContraintAction = "SET NULL"
	ContraintActionSetDefault ContraintAction = "SET DEFAULT"
	ContraintActionCascade    ContraintAction = "CASCADE"
	ContraintActionRestrict   ContraintAction = "RESTRICT"
	ContraintActionNoAction   ContraintAction = "NO ACTION"
)

type Constraint struct {
	Name      string
	Type      ConstraintType
	Target    []string
	Reference *ConstraintReference
	OnDelete  ContraintAction
	OnUpdate  ContraintAction
}

func (c *Constraint) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.Type)
}

type IndexAlgorithm string

var (
	IndexAlgorithmBTree IndexAlgorithm = "btree"
)

type Index struct {
	Name      string
	Algorithm IndexAlgorithm
	Columns   []string
}
