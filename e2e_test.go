package main

import (
	"path/filepath"
	"strings"
	"stijntratsaertit/terramigrate/migration"
	"stijntratsaertit/terramigrate/objects"
	"stijntratsaertit/terramigrate/state"
	"testing"
)

func loadExample(t *testing.T, name string) []*objects.Namespace {
	t.Helper()
	req, err := state.LoadYAML(filepath.Join("examples", name))
	if err != nil {
		t.Fatalf("could not load %s: %v", name, err)
	}
	return req.Namespaces
}

func validateNamespaces(t *testing.T, namespaces []*objects.Namespace) {
	t.Helper()
	for _, ns := range namespaces {
		if err := ns.Valid(); err != nil {
			t.Fatalf("validation failed for namespace %s: %v", ns.Name, err)
		}
	}
}

func diffActions(t *testing.T, existing, desired []*objects.Namespace) []string {
	t.Helper()
	migrators := state.Compare(existing, desired)
	var all []string
	for _, m := range migrators {
		all = append(all, m.GetActions()...)
	}
	return all
}

func assertContainsE2E(t *testing.T, actions []string, substr string) {
	t.Helper()
	for _, a := range actions {
		if strings.Contains(a, substr) {
			return
		}
	}
	t.Errorf("expected actions to contain %q\ngot:\n  %s", substr, strings.Join(actions, "\n  "))
}

func assertNotContains(t *testing.T, actions []string, substr string) {
	t.Helper()
	for _, a := range actions {
		if strings.Contains(a, substr) {
			t.Errorf("expected actions NOT to contain %q, but found: %s", substr, a)
			return
		}
	}
}

// --- Simple example ---

func TestE2E_Simple_FreshDeploy(t *testing.T) {
	desired := loadExample(t, "simple.yaml")
	validateNamespaces(t, desired)

	actions := diffActions(t, nil, desired)

	assertContainsE2E(t, actions, "CREATE TABLE public.users")
	assertContainsE2E(t, actions, "ADD COLUMN id")
	assertContainsE2E(t, actions, "ADD COLUMN email")
	assertContainsE2E(t, actions, "ADD COLUMN name")
	assertContainsE2E(t, actions, "ADD COLUMN created_at")
	assertContainsE2E(t, actions, "ADD CONSTRAINT users_pkey PRIMARY KEY")
	assertContainsE2E(t, actions, "ADD CONSTRAINT users_email_unique UNIQUE")
	assertContainsE2E(t, actions, "CREATE UNIQUE INDEX idx_users_email")
	assertContainsE2E(t, actions, "CREATE SEQUENCE public.users_id_seq")
}

func TestE2E_Simple_NoopWhenIdentical(t *testing.T) {
	desired := loadExample(t, "simple.yaml")
	actions := diffActions(t, desired, desired)

	if len(actions) != 0 {
		t.Errorf("expected no actions when existing == desired, got %d:\n  %s",
			len(actions), strings.Join(actions, "\n  "))
	}
}

func TestE2E_Simple_AddColumn(t *testing.T) {
	existing := loadExample(t, "simple.yaml")

	desired := loadExample(t, "simple.yaml")
	desired[0].Tables[0].Columns = append(desired[0].Tables[0].Columns, &objects.Column{
		Name: "bio", Type: "TEXT", Nullable: true,
	})

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "ADD COLUMN bio")
	assertNotContains(t, actions, "CREATE TABLE")
}

func TestE2E_Simple_DropColumn(t *testing.T) {
	existing := loadExample(t, "simple.yaml")
	desired := loadExample(t, "simple.yaml")
	desired[0].Tables[0].Columns = desired[0].Tables[0].Columns[:2]

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "DROP COLUMN name")
	assertContainsE2E(t, actions, "DROP COLUMN created_at")
}

func TestE2E_Simple_ChangeColumnType(t *testing.T) {
	existing := loadExample(t, "simple.yaml")
	desired := loadExample(t, "simple.yaml")
	desired[0].Tables[0].Columns[0].Type = "BIGINT"

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "ALTER COLUMN id TYPE BIGINT")
}

func TestE2E_Simple_FullTeardown(t *testing.T) {
	existing := loadExample(t, "simple.yaml")
	actions := diffActions(t, existing, nil)

	assertContainsE2E(t, actions, "DROP SCHEMA public")
}

// --- Blog example ---

func TestE2E_Blog_FreshDeploy(t *testing.T) {
	desired := loadExample(t, "blog.yaml")
	validateNamespaces(t, desired)

	actions := diffActions(t, nil, desired)

	assertContainsE2E(t, actions, "CREATE TABLE public.users")
	assertContainsE2E(t, actions, "CREATE TABLE public.posts")
	assertContainsE2E(t, actions, "CREATE TABLE public.comments")
	assertContainsE2E(t, actions, "CREATE SEQUENCE public.users_id_seq")
	assertContainsE2E(t, actions, "CREATE SEQUENCE public.posts_id_seq")
	assertContainsE2E(t, actions, "CREATE SEQUENCE public.comments_id_seq")
	assertContainsE2E(t, actions, "CONSTRAINT posts_author_fk FOREIGN KEY")
	assertContainsE2E(t, actions, "REFERENCES users (id)")
	assertContainsE2E(t, actions, "ON DELETE CASCADE")
}

func TestE2E_Blog_AddNewTable(t *testing.T) {
	existing := loadExample(t, "blog.yaml")
	desired := loadExample(t, "blog.yaml")

	desired[0].Tables = append(desired[0].Tables, &objects.Table{
		Name: "tags",
		Columns: []*objects.Column{
			{Name: "id", Type: "INTEGER", Nullable: false, Default: "nextval('tags_id_seq')", IsPrimaryKey: true},
			{Name: "name", Type: "CHARACTER VARYING", MaxLength: 50, Nullable: false, Default: "''"},
		},
		Constraints: []*objects.Constraint{
			{Name: "tags_pkey", Type: "PRIMARY KEY", Targets: []string{"id"}},
		},
	})

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "CREATE TABLE public.tags")
	assertContainsE2E(t, actions, "ADD COLUMN id")
	assertContainsE2E(t, actions, "ADD COLUMN name")
	assertNotContains(t, actions, "CREATE TABLE public.users")
	assertNotContains(t, actions, "CREATE TABLE public.posts")
}

func TestE2E_Blog_DropTable(t *testing.T) {
	existing := loadExample(t, "blog.yaml")
	desired := loadExample(t, "blog.yaml")
	desired[0].Tables = desired[0].Tables[:2]

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "DROP TABLE public.comments")
	assertNotContains(t, actions, "DROP TABLE public.users")
	assertNotContains(t, actions, "DROP TABLE public.posts")
}

func TestE2E_Blog_ModifyConstraint(t *testing.T) {
	existing := loadExample(t, "blog.yaml")
	desired := loadExample(t, "blog.yaml")

	for _, c := range desired[0].Tables[1].Constraints {
		if c.Name == "posts_author_fk" {
			c.OnDelete = "SET NULL"
		}
	}

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "DROP CONSTRAINT posts_author_fk")
	assertContainsE2E(t, actions, "ADD CONSTRAINT posts_author_fk")
	assertContainsE2E(t, actions, "ON DELETE SET NULL")
}

func TestE2E_Blog_AddIndex(t *testing.T) {
	existing := loadExample(t, "blog.yaml")
	desired := loadExample(t, "blog.yaml")

	desired[0].Tables[2].Indices = append(desired[0].Tables[2].Indices, &objects.Index{
		Name: "idx_comments_user", Algorithm: "btree", Columns: []string{"user_id"},
	})

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "CREATE INDEX idx_comments_user")
	assertContainsE2E(t, actions, "user_id")
}

// --- Ecommerce example ---

func TestE2E_Ecommerce_FreshDeploy(t *testing.T) {
	desired := loadExample(t, "ecommerce.yaml")
	validateNamespaces(t, desired)

	actions := diffActions(t, nil, desired)

	assertContainsE2E(t, actions, "CREATE SCHEMA public")
	assertContainsE2E(t, actions, "CREATE TABLE public.customers")
	assertContainsE2E(t, actions, "CREATE TABLE public.products")
	assertContainsE2E(t, actions, "CREATE TABLE public.orders")
	assertContainsE2E(t, actions, "CREATE TABLE public.order_items")
	assertContainsE2E(t, actions, "CREATE SCHEMA analytics")
	assertContainsE2E(t, actions, "CREATE TABLE analytics.page_views")
	assertContainsE2E(t, actions, "ON DELETE RESTRICT")
	assertContainsE2E(t, actions, "ON DELETE CASCADE")
}

func TestE2E_Ecommerce_DropSchema(t *testing.T) {
	existing := loadExample(t, "ecommerce.yaml")
	desired := loadExample(t, "ecommerce.yaml")
	desired = desired[:1]

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "DROP SCHEMA analytics")
	assertNotContains(t, actions, "DROP SCHEMA public")
}

func TestE2E_Ecommerce_AddColumnToExisting(t *testing.T) {
	existing := loadExample(t, "ecommerce.yaml")
	desired := loadExample(t, "ecommerce.yaml")

	for _, tbl := range desired[0].Tables {
		if tbl.Name == "products" {
			tbl.Columns = append(tbl.Columns, &objects.Column{
				Name: "weight_grams", Type: "INTEGER", Nullable: true,
			})
		}
	}

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "ADD COLUMN weight_grams")
	assertNotContains(t, actions, "CREATE TABLE")
}

func TestE2E_Ecommerce_ChangeSequenceType(t *testing.T) {
	existing := loadExample(t, "ecommerce.yaml")
	desired := loadExample(t, "ecommerce.yaml")

	for _, seq := range desired[0].Sequences {
		if seq.Name == "orders_id_seq" {
			seq.Type = "integer"
		}
	}

	actions := diffActions(t, existing, desired)
	assertContainsE2E(t, actions, "ALTER SEQUENCE public.orders_id_seq AS integer")
}

// --- Migration file generation e2e ---

func TestE2E_MigrationFileGeneration(t *testing.T) {
	desired := loadExample(t, "simple.yaml")
	actions := diffActions(t, nil, desired)

	upSQL := strings.Join(actions, "\n")
	downSQL := migration.GenerateDownSQL(actions, nil)

	m := migration.NewMigration("initial schema", upSQL, downSQL)

	dir := t.TempDir()
	if err := m.Write(dir); err != nil {
		t.Fatalf("could not write migration: %v", err)
	}

	loaded, err := migration.LoadMigration(filepath.Join(dir, m.DirName()))
	if err != nil {
		t.Fatalf("could not reload migration: %v", err)
	}

	if loaded.UpSQL != upSQL {
		t.Error("loaded up SQL does not match written up SQL")
	}
	if loaded.DownSQL != downSQL {
		t.Error("loaded down SQL does not match written down SQL")
	}
	if !loaded.VerifyChecksum() {
		t.Error("checksum verification failed on reloaded migration")
	}
}

// --- Down migration correctness ---

func TestE2E_DownMigration_FreshSimple(t *testing.T) {
	desired := loadExample(t, "simple.yaml")
	actions := diffActions(t, nil, desired)
	downSQL := migration.GenerateDownSQL(actions, nil)

	assertContainsE2E(t, strings.Split(downSQL, "\n"), "DROP SEQUENCE public.users_id_seq;")
	assertContainsE2E(t, strings.Split(downSQL, "\n"), "DROP INDEX idx_users_email;")
	assertContainsE2E(t, strings.Split(downSQL, "\n"), "DROP CONSTRAINT users_email_unique;")
	assertContainsE2E(t, strings.Split(downSQL, "\n"), "DROP CONSTRAINT users_pkey;")
	assertContainsE2E(t, strings.Split(downSQL, "\n"), "DROP TABLE public.users;")
}

func TestE2E_DownMigration_ColumnEdit(t *testing.T) {
	existing := loadExample(t, "simple.yaml")
	desired := loadExample(t, "simple.yaml")
	desired[0].Tables[0].Columns[0].Type = "BIGINT"

	actions := diffActions(t, existing, desired)

	es := &migration.ExistingState{
		ColumnTypes:    map[string]string{"public.users.id": "INTEGER"},
		ColumnDefaults: map[string]string{},
		ColumnNullable: map[string]bool{},
		SequenceTypes:  map[string]string{},
	}
	downSQL := migration.GenerateDownSQL(actions, es)

	assertContainsE2E(t, strings.Split(downSQL, "\n"), "TYPE INTEGER")
}

func TestE2E_DownMigration_ReversedOrder(t *testing.T) {
	desired := loadExample(t, "blog.yaml")
	actions := diffActions(t, nil, desired)
	downSQL := migration.GenerateDownSQL(actions, nil)
	lines := strings.Split(downSQL, "\n")

	seqIdx := -1
	tableIdx := -1
	for i, l := range lines {
		if strings.Contains(l, "DROP SEQUENCE public.users_id_seq") {
			seqIdx = i
		}
		if strings.Contains(l, "DROP TABLE public.users") {
			tableIdx = i
		}
	}

	if seqIdx == -1 || tableIdx == -1 {
		t.Fatalf("missing expected down actions in:\n%s", downSQL)
	}
	if seqIdx >= tableIdx {
		t.Errorf("sequence drop (line %d) should come before table drop (line %d) in reversed order", seqIdx, tableIdx)
	}
}

// --- Export roundtrip ---

func TestE2E_ExportRoundtrip(t *testing.T) {
	original := loadExample(t, "simple.yaml")

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "exported.yaml")

	s := &state.State{Database: &objects.Database{Name: "test", Namespaces: original}}
	if err := s.ExportYAML(exportPath); err != nil {
		t.Fatalf("could not export: %v", err)
	}

	reloaded, err := state.LoadYAML(exportPath)
	if err != nil {
		t.Fatalf("could not reload exported YAML: %v", err)
	}

	actions := diffActions(t, original, reloaded.Namespaces)
	if len(actions) != 0 {
		t.Errorf("expected no diff after export roundtrip, got %d actions:\n  %s",
			len(actions), strings.Join(actions, "\n  "))
	}
}
