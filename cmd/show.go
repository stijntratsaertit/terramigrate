package cmd

import (
	"fmt"
	"stijntratsaertit/terramigrate/database/generic"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the state",
	RunE:  show,
}

func show(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Warningf("could not connect to database: %v", err)
		return err
	}

	state := db.GetState()
	fmt.Print(state.String())

	return nil
}
