package cmd

import (
	"stijntratsaertit/terramigrate/database/generic"
	"stijntratsaertit/terramigrate/state"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	applyCmd.Flags().StringVar(&applyFile, "file", "./db.yaml", "The path to the state to apply")
	applyCmd.Flags().BoolVar(&applyForce, "force", false, "The path to the state to apply")
	rootCmd.AddCommand(applyCmd)
}

var (
	applyFile  string
	applyForce bool
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the state",
	RunE:  apply,
}

func apply(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Errorf("could not connect to database: %v", err)
		return err
	}
	s := db.GetState()

	req, err := state.LoadYAML(applyFile)
	if err != nil {
		return err
	}

	for _, namespace := range req.Namespaces {
		if err := namespace.Valid(); err != nil {
			return err
		}
	}

	migrators := state.Compare(req.Namespaces, s.Database.Namespaces)
	if len(migrators) == 0 {
		log.Info("no differences found")
		return nil
	}

	for _, migrator := range migrators {
		log.Infof("difference found: %s", migrator.String())
		if err := db.ExecuteTransaction(migrator); err != nil {
			return err
		}
	}
	return nil
}
