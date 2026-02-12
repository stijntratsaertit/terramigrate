package cmd

import (
	"fmt"
	"stijntratsaertit/terramigrate/database/generic"
	"stijntratsaertit/terramigrate/migration"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "Number of migrations to rollback")
	rollbackCmd.Flags().StringVar(&rollbackMigrationsDir, "migrations-dir", "./migrations", "The migrations directory")
	rootCmd.AddCommand(rollbackCmd)
}

var (
	rollbackSteps         int
	rollbackMigrationsDir string
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the last N applied migrations",
	RunE:  rollback,
}

func rollback(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Errorf("could not connect to database: %v", err)
		return err
	}

	applied, err := migration.GetAppliedMigrationsFromDisk(db, rollbackMigrationsDir)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		fmt.Println("No applied migrations to rollback.")
		return nil
	}

	if rollbackSteps > len(applied) {
		rollbackSteps = len(applied)
	}

	toRollback := applied[len(applied)-rollbackSteps:]

	fmt.Printf("Rolling back %d migration(s):\n", len(toRollback))
	for i := len(toRollback) - 1; i >= 0; i-- {
		m := toRollback[i]
		fmt.Printf("  Rolling back %s... ", m.DirName())
		if err := migration.RollbackMigration(db, m); err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")
	}

	fmt.Printf("\nSuccessfully rolled back %d migration(s).\n", len(toRollback))
	return nil
}
