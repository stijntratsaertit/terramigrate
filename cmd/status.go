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
	statusCmd.Flags().StringVar(&statusMigrationsDir, "migrations-dir", "./migrations", "The migrations directory")
	rootCmd.AddCommand(statusCmd)
}

var statusMigrationsDir string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	RunE:  status,
}

func status(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Errorf("could not connect to database: %v", err)
		return err
	}

	statuses, err := migration.GetMigrationStatuses(db, statusMigrationsDir)
	if err != nil {
		return err
	}

	if len(statuses) == 0 {
		fmt.Println("No migrations found.")
		return nil
	}

	fmt.Printf("%-30s %-10s %s\n", "VERSION", "STATUS", "DESCRIPTION")
	fmt.Println("-------------------------------------------------------------------")

	for _, s := range statuses {
		statusStr := "pending"
		if s.Applied {
			statusStr = "applied"
		}
		if s.Drift {
			statusStr = "DRIFT"
		}
		fmt.Printf("%-30s %-10s %s\n", s.Migration.Version, statusStr, s.Migration.Description)
	}

	return nil
}
