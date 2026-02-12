package objects

import (
	"strings"
	"testing"
)

func TestColumn_Valid_NoName(t *testing.T) {
	c := &Column{Type: "INTEGER"}
	err := c.Valid()
	if err == nil {
		t.Error("expected error for column with no name")
	}
}

func TestColumn_Valid_NoType(t *testing.T) {
	c := &Column{Name: "id"}
	err := c.Valid()
	if err == nil {
		t.Error("expected error for column with no type")
	}
}

func TestColumn_Valid_VarcharNoMaxLength(t *testing.T) {
	c := &Column{Name: "email", Type: "CHARACTER VARYING", Nullable: true}
	err := c.Valid()
	if err == nil {
		t.Error("expected error for varchar with no max length")
	}
}

func TestColumn_Valid_NonVarcharWithMaxLength(t *testing.T) {
	c := &Column{Name: "id", Type: "INTEGER", MaxLength: 10, Nullable: true}
	err := c.Valid()
	if err == nil {
		t.Error("expected error for non-varchar with max length")
	}
}

func TestColumn_Valid_NotNullableNoDefault(t *testing.T) {
	c := &Column{Name: "name", Type: "TEXT", Nullable: false}
	err := c.Valid()
	if err == nil {
		t.Error("expected error for non-nullable column with no default and not PK")
	}
	if !strings.Contains(err.Error(), "not nullable") {
		t.Errorf("error message should mention 'not nullable', got: %s", err.Error())
	}
}

func TestColumn_Valid_PrimaryKeyOK(t *testing.T) {
	c := &Column{Name: "id", Type: "INTEGER", Nullable: false, IsPrimaryKey: true}
	err := c.Valid()
	if err != nil {
		t.Errorf("expected PK column to be valid, got: %v", err)
	}
}

func TestColumn_Valid_NullableOK(t *testing.T) {
	c := &Column{Name: "notes", Type: "TEXT", Nullable: true}
	err := c.Valid()
	if err != nil {
		t.Errorf("expected nullable column to be valid, got: %v", err)
	}
}

func TestColumn_Valid_WithDefaultOK(t *testing.T) {
	c := &Column{Name: "status", Type: "TEXT", Nullable: false, Default: "'active'"}
	err := c.Valid()
	if err != nil {
		t.Errorf("expected column with default to be valid, got: %v", err)
	}
}

func TestTable_Valid_NoName(t *testing.T) {
	tbl := &Table{}
	err := tbl.Valid()
	if err == nil {
		t.Error("expected error for table with no name")
	}
}

func TestTable_Valid_NameTooLong(t *testing.T) {
	tbl := &Table{Name: strings.Repeat("a", 64)}
	err := tbl.Valid()
	if err == nil {
		t.Error("expected error for table with name > 63 chars")
	}
}

func TestSequence_Valid_NoName(t *testing.T) {
	s := &Sequence{Type: "bigint"}
	err := s.Valid()
	if err == nil {
		t.Error("expected error for sequence with no name")
	}
}

func TestSequence_Valid_UnsupportedType(t *testing.T) {
	s := &Sequence{Name: "my_seq", Type: "smallint"}
	err := s.Valid()
	if err == nil {
		t.Error("expected error for unsupported sequence type")
	}
}

func TestSequence_Valid_OK(t *testing.T) {
	s := &Sequence{Name: "my_seq", Type: "bigint"}
	err := s.Valid()
	if err != nil {
		t.Errorf("expected sequence to be valid, got: %v", err)
	}
}

func TestNamespace_Valid_PropagatesTableError(t *testing.T) {
	ns := &Namespace{
		Name: "public",
		Tables: []*Table{
			{Name: ""},
		},
	}
	err := ns.Valid()
	if err == nil {
		t.Error("expected namespace validation to catch table error")
	}
}
