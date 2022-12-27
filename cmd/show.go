package cmd

import (
	"fmt"
	"stijntratsaertit/terramigrate/config"
	"stijntratsaertit/terramigrate/database"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the state",
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
