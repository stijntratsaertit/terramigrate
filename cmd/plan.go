package cmd

import (
	"stijntratsaertit/terramigrate/parser"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	rootCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan the state",
	RunE: func(cmd *cobra.Command, args []string) error {
		request, err := parser.LoadYAML("v2.yaml")
		if err != nil {
			return err
		}

		log.Infof("request: %+v", request)

		return nil
	},
}
