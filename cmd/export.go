package cmd

import (
	"stijntratsaertit/terramigrate/database/generic"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func init() {
	exportCmd.Flags().StringVarP(&exportFile, "file", "f", "./db.yaml", "The path to export the state to")
	rootCmd.AddCommand(exportCmd)
}

var exportFile string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the state",
	RunE:  export,
}

func export(cmd *cobra.Command, args []string) error {
	db, err := generic.GetDatabaseAdapter(viper.GetString("adapter"))
	if err != nil {
		log.Warningf("could not connect to database: %v", err)
		return err
	}

	if db.GetState().ExportYAML(exportFile) != nil {
		return err
	}

	return nil
}
