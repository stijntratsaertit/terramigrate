package cmd

import (
	"stijntratsaertit/terramigrate/database/generic"
	"stijntratsaertit/terramigrate/state"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	planCmd.Flags().StringVar(&planFile, "file", "./db.yaml", "The path to the state to plan")
	rootCmd.AddCommand(planCmd)
}

var (
	planFile string
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan the state",
	RunE:  plan,
}

func plan(cmd *cobra.Command, args []string) (err error) {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Errorf("could not connect to database: %v", err)
		return
	}
	s := db.GetState()

	req, err := state.LoadYAML(planFile)

	migrators := state.Compare(req.Namespaces, s.Database.Namespaces)
	if len(migrators) == 0 {
		log.Info("No differences found")
		return
	}

	for _, migrator := range migrators {
		actions := migrator.GetActions()
		log.Info(migrator.String())
		for _, action := range actions {
			log.Infof("Action: %s", action)
		}
	}
	return
}
