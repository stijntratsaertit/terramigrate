package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"stijntratsaertit/terramigrate/database/generic"
	"stijntratsaertit/terramigrate/migration"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	applyCmd.Flags().BoolVar(&applyAutoApprove, "auto-approve", false, "Skip interactive confirmation (for CI)")
	applyCmd.Flags().StringVar(&applyMigrationsDir, "migrations-dir", "./migrations", "The migrations directory")
	rootCmd.AddCommand(applyCmd)
}

var (
	applyAutoApprove  bool
	applyMigrationsDir string
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply pending migrations",
	RunE:  apply,
}

func apply(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Errorf("could not connect to database: %v", err)
		return err
	}

	pending, err := migration.GetPendingMigrations(db, applyMigrationsDir)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		fmt.Println("No pending migrations to apply.")
		return nil
	}

	fmt.Printf("Pending migrations (%d):\n\n", len(pending))
	for _, m := range pending {
		fmt.Printf("  %s\n", m.DirName())
		for _, line := range strings.Split(m.UpSQL, "\n") {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println()
	}

	if !applyAutoApprove {
		fmt.Print("Proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	for _, m := range pending {
		fmt.Printf("Applying %s... ", m.DirName())
		if err := migration.ApplyMigration(db, m); err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")
	}

	fmt.Printf("\nSuccessfully applied %d migration(s).\n", len(pending))
	return nil
}
