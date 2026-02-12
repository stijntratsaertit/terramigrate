package adapter

import (
	"stijntratsaertit/terramigrate/state"
	"time"
)

type AppliedMigration struct {
	Version     string
	Description string
	AppliedAt   time.Time
	Checksum    string
}

type Adapter interface {
	GetState() *state.State
	LoadState() error
	ExecuteTransaction(*state.Migrator) error
	ExecuteSQL(sql string) error
	EnsureMigrationTable() error
	RecordMigration(version, description, checksum string) error
	RemoveMigration(version string) error
	GetAppliedMigrations() ([]AppliedMigration, error)
}
