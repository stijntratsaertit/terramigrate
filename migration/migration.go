package migration

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Migration struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Checksum    string `yaml:"checksum"`
	CreatedAt   string `yaml:"created_at"`
	UpSQL       string `yaml:"-"`
	DownSQL     string `yaml:"-"`
}

func NewMigration(description string, upSQL, downSQL string) *Migration {
	version := time.Now().Format("20060102_150405")
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(upSQL)))

	return &Migration{
		Version:     version,
		Description: sanitizeDescription(description),
		Checksum:    checksum,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpSQL:       upSQL,
		DownSQL:     downSQL,
	}
}

func sanitizeDescription(desc string) string {
	desc = strings.ToLower(desc)
	desc = strings.ReplaceAll(desc, " ", "_")
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, desc)
	if len(safe) > 60 {
		safe = safe[:60]
	}
	return safe
}

func (m *Migration) DirName() string {
	return fmt.Sprintf("%s_%s", m.Version, m.Description)
}

func (m *Migration) Write(migrationsDir string) error {
	dir := filepath.Join(migrationsDir, m.DirName())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not create migration directory %s: %v", dir, err)
	}

	if err := os.WriteFile(filepath.Join(dir, "up.sql"), []byte(m.UpSQL), 0644); err != nil {
		return fmt.Errorf("could not write up.sql: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "down.sql"), []byte(m.DownSQL), 0644); err != nil {
		return fmt.Errorf("could not write down.sql: %v", err)
	}

	planData, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("could not marshal plan.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plan.yaml"), planData, 0644); err != nil {
		return fmt.Errorf("could not write plan.yaml: %v", err)
	}

	return nil
}

func LoadMigration(dir string) (*Migration, error) {
	planPath := filepath.Join(dir, "plan.yaml")
	planData, err := os.ReadFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("could not read plan.yaml in %s: %v", dir, err)
	}

	m := &Migration{}
	if err := yaml.Unmarshal(planData, m); err != nil {
		return nil, fmt.Errorf("could not parse plan.yaml in %s: %v", dir, err)
	}

	upData, err := os.ReadFile(filepath.Join(dir, "up.sql"))
	if err != nil {
		return nil, fmt.Errorf("could not read up.sql in %s: %v", dir, err)
	}
	m.UpSQL = string(upData)

	downData, err := os.ReadFile(filepath.Join(dir, "down.sql"))
	if err != nil {
		return nil, fmt.Errorf("could not read down.sql in %s: %v", dir, err)
	}
	m.DownSQL = string(downData)

	return m, nil
}

func LoadAllMigrations(migrationsDir string) ([]*Migration, error) {
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read migrations directory: %v", err)
	}

	var migrations []*Migration
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		m, err := LoadMigration(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migration) VerifyChecksum() bool {
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(m.UpSQL)))
	return expected == m.Checksum
}
