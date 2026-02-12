package cmd

import (
	"fmt"
	"strings"
	"stijntratsaertit/terramigrate/database/generic"
	"stijntratsaertit/terramigrate/migration"
	"stijntratsaertit/terramigrate/objects"
	"stijntratsaertit/terramigrate/state"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	planCmd.Flags().StringVar(&planFile, "file", "./db.yaml", "The path to the desired state YAML")
	planCmd.Flags().StringVar(&planDescription, "description", "", "Short description for the migration")
	planCmd.Flags().StringVar(&planMigrationsDir, "migrations-dir", "./migrations", "The migrations directory")
	rootCmd.AddCommand(planCmd)
}

var (
	planFile          string
	planDescription   string
	planMigrationsDir string
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan the migration and generate migration files",
	RunE:  plan,
}

func plan(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Errorf("could not connect to database: %v", err)
		return err
	}
	s := db.GetState()

	req, err := state.LoadYAML(planFile)
	if err != nil {
		return err
	}

	for _, namespace := range req.Namespaces {
		if err := namespace.Valid(); err != nil {
			return err
		}
	}

	migrators := state.Compare(s.Database.Namespaces, req.Namespaces)

	var allActions []string
	for _, m := range migrators {
		allActions = append(allActions, m.GetActions()...)
	}

	if len(allActions) == 0 {
		log.Info("no differences found, nothing to plan")
		return nil
	}

	existingState := buildExistingStateFromNamespaces(s.Database.Namespaces)
	upSQL := strings.Join(allActions, "\n")
	downSQL := migration.GenerateDownSQL(allActions, existingState)

	if planDescription == "" {
		planDescription = generateDescription(allActions)
	}

	m := migration.NewMigration(planDescription, upSQL, downSQL)

	if err := m.Write(planMigrationsDir); err != nil {
		return fmt.Errorf("could not write migration: %v", err)
	}

	fmt.Printf("Migration planned: %s\n\n", m.DirName())
	fmt.Println("--- UP (forward) ---")
	fmt.Println(upSQL)
	fmt.Println()
	fmt.Println("--- DOWN (rollback) ---")
	fmt.Println(downSQL)
	fmt.Printf("\nMigration files written to %s/%s/\n", planMigrationsDir, m.DirName())

	return nil
}

func buildExistingStateFromNamespaces(namespaces []*objects.Namespace) *migration.ExistingState {
	es := &migration.ExistingState{
		ColumnTypes:    make(map[string]string),
		ColumnDefaults: make(map[string]string),
		ColumnNullable: make(map[string]bool),
		SequenceTypes:  make(map[string]string),
	}

	for _, ns := range namespaces {
		for _, t := range ns.Tables {
			fullName := fmt.Sprintf("%s.%s", ns.Name, t.Name)
			for _, col := range t.Columns {
				key := fullName + "." + col.Name
				es.ColumnTypes[key] = col.Type
				es.ColumnDefaults[key] = col.Default
				es.ColumnNullable[key] = col.Nullable
			}
		}
		for _, seq := range ns.Sequences {
			fullName := fmt.Sprintf("%s.%s", ns.Name, seq.Name)
			es.SequenceTypes[fullName] = seq.Type
		}
	}

	return es
}

func generateDescription(actions []string) string {
	if len(actions) == 0 {
		return "empty_migration"
	}

	first := strings.ToLower(actions[0])
	switch {
	case strings.Contains(first, "create table"):
		return "schema_changes"
	case strings.Contains(first, "alter table"):
		return "table_alterations"
	case strings.Contains(first, "create schema"):
		return "create_schema"
	default:
		return "migration"
	}
}
