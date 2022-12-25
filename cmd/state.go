package cmd

import (
	"fmt"
	"stijntratsaertit/terramigrate/config"
	"stijntratsaertit/terramigrate/database"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(stateListCmd)
}

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Commands to manage the state",
}

var stateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the state",
	RunE: func(cmd *cobra.Command, args []string) error {

		db, err := database.GetDatabase(config.Get().DatabaseConnectionParams())
		if err != nil {
			log.Warningf("could not connect to database: %v", err)
			return err
		}

		state := db.GetState()
		fmt.Print(state.String())

		return nil
	},
}
