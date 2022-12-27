package cmd

import (
	"stijntratsaertit/terramigrate/config"
	"stijntratsaertit/terramigrate/database"
	"stijntratsaertit/terramigrate/parser"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	planCmd.Flags().StringVarP(&targetFile, "file", "f", "./db.yaml", "The path to export the state to")
	rootCmd.AddCommand(exportCmd)
}

var targetFile string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the state",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.GetDatabase(config.Get().DatabaseConnectionParams())
		if err != nil {
			log.Warningf("could not connect to database: %v", err)
			return err
		}

		state := db.GetState()

		request := &parser.Request{Namespaces: state.Database.Namespaces}
		if parser.ExportYAML(targetFile, request) != nil {
			return err
		}

		return nil
	},
}
