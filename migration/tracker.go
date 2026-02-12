package migration

import (
	"fmt"
	"stijntratsaertit/terramigrate/database/adapter"
)

type MigrationStatus struct {
	Migration *Migration
	Applied   bool
	Checksum  string
	Drift     bool
}

func GetPendingMigrations(db adapter.Adapter, migrationsDir string) ([]*Migration, error) {
	if err := db.EnsureMigrationTable(); err != nil {
		return nil, err
	}

	applied, err := db.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	allMigrations, err := LoadAllMigrations(migrationsDir)
	if err != nil {
		return nil, err
	}

	appliedSet := make(map[string]bool)
	for _, m := range applied {
		appliedSet[m.Version] = true
	}

	var pending []*Migration
	for _, m := range allMigrations {
		if !appliedSet[m.Version] {
			pending = append(pending, m)
		}
	}

	return pending, nil
}

func GetMigrationStatuses(db adapter.Adapter, migrationsDir string) ([]MigrationStatus, error) {
	if err := db.EnsureMigrationTable(); err != nil {
		return nil, err
	}

	applied, err := db.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	allMigrations, err := LoadAllMigrations(migrationsDir)
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]adapter.AppliedMigration)
	for _, m := range applied {
		appliedMap[m.Version] = m
	}

	var statuses []MigrationStatus
	for _, m := range allMigrations {
		status := MigrationStatus{Migration: m}
		if am, ok := appliedMap[m.Version]; ok {
			status.Applied = true
			status.Checksum = am.Checksum
			status.Drift = am.Checksum != m.Checksum
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func ApplyMigration(db adapter.Adapter, m *Migration) error {
	if !m.VerifyChecksum() {
		return fmt.Errorf("migration %s has been modified since it was planned (checksum mismatch)", m.Version)
	}

	if err := db.ExecuteSQL(m.UpSQL); err != nil {
		return fmt.Errorf("failed to apply migration %s: %v", m.Version, err)
	}

	if err := db.RecordMigration(m.Version, m.Description, m.Checksum); err != nil {
		return fmt.Errorf("migration %s applied but could not record it: %v", m.Version, err)
	}

	return nil
}

func RollbackMigration(db adapter.Adapter, m *Migration) error {
	if err := db.ExecuteSQL(m.DownSQL); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %v", m.Version, err)
	}

	if err := db.RemoveMigration(m.Version); err != nil {
		return fmt.Errorf("migration %s rolled back but could not remove record: %v", m.Version, err)
	}

	return nil
}

func GetAppliedMigrationsFromDisk(db adapter.Adapter, migrationsDir string) ([]*Migration, error) {
	if err := db.EnsureMigrationTable(); err != nil {
		return nil, err
	}

	applied, err := db.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	allMigrations, err := LoadAllMigrations(migrationsDir)
	if err != nil {
		return nil, err
	}

	appliedSet := make(map[string]bool)
	for _, m := range applied {
		appliedSet[m.Version] = true
	}

	var result []*Migration
	for _, m := range allMigrations {
		if appliedSet[m.Version] {
			result = append(result, m)
		}
	}

	return result, nil
}
