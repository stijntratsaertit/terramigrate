package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewMigration_SetsVersion(t *testing.T) {
	m := NewMigration("test migration", "CREATE TABLE foo;", "DROP TABLE foo;")
	if m.Version == "" {
		t.Error("expected version to be set")
	}
	if m.Description != "test_migration" {
		t.Errorf("expected description 'test_migration', got '%s'", m.Description)
	}
}

func TestMigration_WriteAndLoad(t *testing.T) {
	dir := t.TempDir()

	upSQL := "CREATE TABLE public.users ();\nALTER TABLE public.users ADD COLUMN id INTEGER NOT NULL DEFAULT 1;"
	downSQL := "ALTER TABLE public.users DROP COLUMN id;\nDROP TABLE public.users;"

	m := NewMigration("add users", upSQL, downSQL)
	if err := m.Write(dir); err != nil {
		t.Fatalf("could not write migration: %v", err)
	}

	migrationDir := filepath.Join(dir, m.DirName())
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		t.Fatalf("migration directory not created: %s", migrationDir)
	}

	loaded, err := LoadMigration(migrationDir)
	if err != nil {
		t.Fatalf("could not load migration: %v", err)
	}

	if loaded.Version != m.Version {
		t.Errorf("version mismatch: %s != %s", loaded.Version, m.Version)
	}
	if loaded.UpSQL != upSQL {
		t.Errorf("up SQL mismatch")
	}
	if loaded.DownSQL != downSQL {
		t.Errorf("down SQL mismatch")
	}
}

func TestLoadAllMigrations_SortedByVersion(t *testing.T) {
	dir := t.TempDir()

	m1 := &Migration{Version: "20260101_100000", Description: "first", UpSQL: "SELECT 1;", DownSQL: "SELECT 1;", Checksum: "abc"}
	m2 := &Migration{Version: "20260102_100000", Description: "second", UpSQL: "SELECT 2;", DownSQL: "SELECT 2;", Checksum: "def"}

	if err := m2.Write(dir); err != nil {
		t.Fatalf("could not write m2: %v", err)
	}
	if err := m1.Write(dir); err != nil {
		t.Fatalf("could not write m1: %v", err)
	}

	migrations, err := LoadAllMigrations(dir)
	if err != nil {
		t.Fatalf("could not load all migrations: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}
	if migrations[0].Version != "20260101_100000" {
		t.Errorf("expected first migration to be 20260101_100000, got %s", migrations[0].Version)
	}
}

func TestLoadAllMigrations_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	migrations, err := LoadAllMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

func TestLoadAllMigrations_NonexistentDir(t *testing.T) {
	migrations, err := LoadAllMigrations("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if migrations != nil {
		t.Errorf("expected nil, got %v", migrations)
	}
}

func TestVerifyChecksum(t *testing.T) {
	m := NewMigration("test", "CREATE TABLE foo;", "DROP TABLE foo;")
	if !m.VerifyChecksum() {
		t.Error("checksum should verify for unmodified migration")
	}

	m.UpSQL = "CREATE TABLE bar;"
	if m.VerifyChecksum() {
		t.Error("checksum should not verify after modifying UpSQL")
	}
}

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Add Users Table", "add_users_table"},
		{"hello world!", "hello_world"},
		{"test-migration", "testmigration"},
		{strings.Repeat("a", 100), strings.Repeat("a", 60)},
	}

	for _, tt := range tests {
		result := sanitizeDescription(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeDescription(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
