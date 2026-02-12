package state

import (
	"strings"
	"stijntratsaertit/terramigrate/objects"
	"testing"
)

func TestCompare_NoChanges(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	for _, m := range migrators {
		if len(m.GetActions()) != 0 {
			t.Errorf("expected no actions, got %d: %v", len(m.GetActions()), m.GetActions())
		}
	}
}

func TestCompare_CreateTable(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "CREATE TABLE public.users")
	assertContains(t, actions, "ADD COLUMN id")
}

func TestCompare_DropTable(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "old_table", Columns: []*objects.Column{}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP TABLE public.old_table")
}

func TestCompare_AddColumn(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
				{Name: "email", Type: "CHARACTER VARYING", MaxLength: 255, Nullable: true},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "ADD COLUMN email")
}

func TestCompare_DropColumn(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
				{Name: "old_col", Type: "TEXT", Nullable: true},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP COLUMN old_col")
}

func TestCompare_AlterColumnType(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "age", Type: "INTEGER", Nullable: true},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "age", Type: "BIGINT", Nullable: true},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "ALTER COLUMN age TYPE BIGINT")
}

func TestCompare_AlterColumnNullable(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "name", Type: "TEXT", Nullable: true},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "name", Type: "TEXT", Nullable: false, Default: "''"},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "SET NOT NULL")
}

func TestCompare_CreateNamespace(t *testing.T) {
	existing := []*objects.Namespace{}
	desired := []*objects.Namespace{
		{Name: "analytics"},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "CREATE SCHEMA analytics")
}

func TestCompare_DropNamespace(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "old_schema"},
	}
	desired := []*objects.Namespace{}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP SCHEMA old_schema")
	for _, m := range migrators {
		if m.IsLocked() {
			return
		}
	}
	t.Error("expected DROP SCHEMA migrator to be locked")
}

func TestCompare_CreateSequence(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public"},
	}
	desired := []*objects.Namespace{
		{Name: "public", Sequences: []*objects.Sequence{
			{Name: "users_id_seq", Type: "bigint"},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "CREATE SEQUENCE public.users_id_seq")
}

func TestCompare_DropSequence(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Sequences: []*objects.Sequence{
			{Name: "old_seq", Type: "bigint"},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public"},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP SEQUENCE public.old_seq")
}

func TestCompare_AlterSequenceType(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Sequences: []*objects.Sequence{
			{Name: "counter_seq", Type: "integer"},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Sequences: []*objects.Sequence{
			{Name: "counter_seq", Type: "bigint"},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "ALTER SEQUENCE public.counter_seq AS bigint")
}

func TestCompare_AddConstraint(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Columns: []*objects.Column{
				{Name: "id", Type: "INTEGER", Nullable: false, Default: "1", IsPrimaryKey: true},
			}, Constraints: []*objects.Constraint{
				{Name: "users_pkey", Type: objects.ConstraintTypePrimaryKey, Targets: []string{"id"}},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "ADD CONSTRAINT users_pkey PRIMARY KEY")
}

func TestCompare_DropConstraint(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Constraints: []*objects.Constraint{
				{Name: "users_pkey", Type: objects.ConstraintTypePrimaryKey, Targets: []string{"id"}},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users"},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP CONSTRAINT users_pkey")
}

func TestCompare_AddIndex(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users"},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Indices: []*objects.Index{
				{Name: "idx_users_email", Unique: true, Algorithm: "btree", Columns: []string{"email"}},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "CREATE UNIQUE INDEX idx_users_email")
}

func TestCompare_DropIndex(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Indices: []*objects.Index{
				{Name: "idx_old", Algorithm: "btree", Columns: []string{"name"}},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users"},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP INDEX public.idx_old")
}

func TestCompare_ModifyIndex(t *testing.T) {
	existing := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Indices: []*objects.Index{
				{Name: "idx_users_email", Unique: false, Algorithm: "btree", Columns: []string{"email"}},
			}},
		}},
	}
	desired := []*objects.Namespace{
		{Name: "public", Tables: []*objects.Table{
			{Name: "users", Indices: []*objects.Index{
				{Name: "idx_users_email", Unique: true, Algorithm: "btree", Columns: []string{"email"}},
			}},
		}},
	}

	migrators := Compare(existing, desired)
	actions := collectActions(migrators)

	assertContains(t, actions, "DROP INDEX public.idx_users_email")
	assertContains(t, actions, "CREATE UNIQUE INDEX idx_users_email")
}

func collectActions(migrators []*Migrator) []string {
	var all []string
	for _, m := range migrators {
		all = append(all, m.GetActions()...)
	}
	return all
}

func assertContains(t *testing.T, actions []string, substr string) {
	t.Helper()
	for _, a := range actions {
		if strings.Contains(a, substr) {
			return
		}
	}
	t.Errorf("expected actions to contain %q, got: %v", substr, actions)
}
