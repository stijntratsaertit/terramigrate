package migration

import (
	"strings"
	"testing"
)

func TestGenerateDownSQL_CreateTable(t *testing.T) {
	up := []string{"CREATE TABLE public.users ();"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP TABLE public.users;") {
		t.Errorf("expected DROP TABLE, got: %s", down)
	}
}

func TestGenerateDownSQL_DropTable(t *testing.T) {
	up := []string{"DROP TABLE public.users;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "WARNING") {
		t.Errorf("expected WARNING for irreversible DROP TABLE, got: %s", down)
	}
}

func TestGenerateDownSQL_AddColumn(t *testing.T) {
	up := []string{"ALTER TABLE public.users ADD COLUMN email TEXT NULL;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP COLUMN email;") {
		t.Errorf("expected DROP COLUMN, got: %s", down)
	}
}

func TestGenerateDownSQL_DropColumn(t *testing.T) {
	up := []string{"ALTER TABLE public.users DROP COLUMN old_col;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "WARNING") {
		t.Errorf("expected WARNING for irreversible DROP COLUMN, got: %s", down)
	}
}

func TestGenerateDownSQL_AlterColumnType_WithState(t *testing.T) {
	up := []string{"ALTER TABLE public.users ALTER COLUMN age TYPE BIGINT;"}
	existing := &ExistingState{
		ColumnTypes:    map[string]string{"public.users.age": "INTEGER"},
		ColumnDefaults: map[string]string{},
		ColumnNullable: map[string]bool{},
		SequenceTypes:  map[string]string{},
	}
	down := GenerateDownSQL(up, existing)

	if !strings.Contains(down, "TYPE INTEGER") {
		t.Errorf("expected restore to INTEGER, got: %s", down)
	}
}

func TestGenerateDownSQL_AlterColumnType_WithoutState(t *testing.T) {
	up := []string{"ALTER TABLE public.users ALTER COLUMN age TYPE BIGINT;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "WARNING") {
		t.Errorf("expected WARNING when original type unknown, got: %s", down)
	}
}

func TestGenerateDownSQL_SetNotNull(t *testing.T) {
	up := []string{"ALTER TABLE public.users ALTER COLUMN name SET NOT NULL;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP NOT NULL") {
		t.Errorf("expected DROP NOT NULL, got: %s", down)
	}
}

func TestGenerateDownSQL_DropNotNull(t *testing.T) {
	up := []string{"ALTER TABLE public.users ALTER COLUMN name DROP NOT NULL;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "SET NOT NULL") {
		t.Errorf("expected SET NOT NULL, got: %s", down)
	}
}

func TestGenerateDownSQL_SetDefault(t *testing.T) {
	up := []string{"ALTER TABLE public.users ALTER COLUMN status SET DEFAULT 'active';"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP DEFAULT") {
		t.Errorf("expected DROP DEFAULT, got: %s", down)
	}
}

func TestGenerateDownSQL_DropDefault_WithState(t *testing.T) {
	up := []string{"ALTER TABLE public.users ALTER COLUMN status DROP DEFAULT;"}
	existing := &ExistingState{
		ColumnTypes:    map[string]string{},
		ColumnDefaults: map[string]string{"public.users.status": "'pending'"},
		ColumnNullable: map[string]bool{},
		SequenceTypes:  map[string]string{},
	}
	down := GenerateDownSQL(up, existing)

	if !strings.Contains(down, "SET DEFAULT 'pending'") {
		t.Errorf("expected restore default, got: %s", down)
	}
}

func TestGenerateDownSQL_CreateSequence(t *testing.T) {
	up := []string{"CREATE SEQUENCE public.users_id_seq;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP SEQUENCE public.users_id_seq;") {
		t.Errorf("expected DROP SEQUENCE, got: %s", down)
	}
}

func TestGenerateDownSQL_DropSequence_WithState(t *testing.T) {
	up := []string{"DROP SEQUENCE public.old_seq;"}
	existing := &ExistingState{
		ColumnTypes:    map[string]string{},
		ColumnDefaults: map[string]string{},
		ColumnNullable: map[string]bool{},
		SequenceTypes:  map[string]string{"public.old_seq": "bigint"},
	}
	down := GenerateDownSQL(up, existing)

	if !strings.Contains(down, "CREATE SEQUENCE public.old_seq AS bigint;") {
		t.Errorf("expected CREATE SEQUENCE with type, got: %s", down)
	}
}

func TestGenerateDownSQL_CreateSchema(t *testing.T) {
	up := []string{"CREATE SCHEMA analytics;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP SCHEMA analytics CASCADE;") {
		t.Errorf("expected DROP SCHEMA, got: %s", down)
	}
}

func TestGenerateDownSQL_DropSchema(t *testing.T) {
	up := []string{"DROP SCHEMA analytics CASCADE;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "WARNING") {
		t.Errorf("expected WARNING for irreversible DROP SCHEMA, got: %s", down)
	}
}

func TestGenerateDownSQL_CreateIndex(t *testing.T) {
	up := []string{"CREATE UNIQUE INDEX idx_email ON public.users USING btree (email);"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP INDEX idx_email;") {
		t.Errorf("expected DROP INDEX, got: %s", down)
	}
}

func TestGenerateDownSQL_DropIndex(t *testing.T) {
	up := []string{"DROP INDEX public.idx_old;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "WARNING") {
		t.Errorf("expected WARNING for irreversible DROP INDEX, got: %s", down)
	}
}

func TestGenerateDownSQL_AddConstraint(t *testing.T) {
	up := []string{"ALTER TABLE public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "DROP CONSTRAINT users_pkey;") {
		t.Errorf("expected DROP CONSTRAINT, got: %s", down)
	}
}

func TestGenerateDownSQL_DropConstraint(t *testing.T) {
	up := []string{"ALTER TABLE public.users DROP CONSTRAINT users_pkey;"}
	down := GenerateDownSQL(up, nil)

	if !strings.Contains(down, "WARNING") {
		t.Errorf("expected WARNING for irreversible DROP CONSTRAINT, got: %s", down)
	}
}

func TestGenerateDownSQL_ReversedOrder(t *testing.T) {
	up := []string{
		"CREATE TABLE public.users ();",
		"ALTER TABLE public.users ADD COLUMN id INTEGER NOT NULL DEFAULT 1;",
	}
	down := GenerateDownSQL(up, nil)
	lines := strings.Split(down, "\n")

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got: %s", down)
	}
	if !strings.Contains(lines[0], "DROP COLUMN id") {
		t.Errorf("first down action should reverse last up action, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "DROP TABLE public.users") {
		t.Errorf("second down action should reverse first up action, got: %s", lines[1])
	}
}

func TestBuildExistingState(t *testing.T) {
	tables := []*ExistingTableInfo{
		{
			FullName: "public.users",
			Columns: []ExistingColumnInfo{
				{Name: "id", Type: "INTEGER", Default: "nextval('users_id_seq')", Nullable: false},
				{Name: "name", Type: "TEXT", Default: "", Nullable: true},
			},
		},
	}

	state := BuildExistingState(tables)

	if state.ColumnTypes["public.users.id"] != "INTEGER" {
		t.Errorf("expected INTEGER, got %s", state.ColumnTypes["public.users.id"])
	}
	if state.ColumnDefaults["public.users.id"] != "nextval('users_id_seq')" {
		t.Errorf("expected default, got %s", state.ColumnDefaults["public.users.id"])
	}
	if state.ColumnNullable["public.users.name"] != true {
		t.Error("expected name to be nullable")
	}
}
