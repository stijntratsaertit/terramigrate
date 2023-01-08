package cmd

import (
	"stijntratsaertit/terramigrate/state"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func init() {
	planCmd.Flags().StringVarP(&planFile, "file", "f", "./db.yaml", "The path to the state to plan")
	rootCmd.AddCommand(planCmd)
}

var planFile string

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan the state",
	RunE:  plan,
}

func plan(cmd *cobra.Command, args []string) error {
	request, err := state.LoadYAML(planFile)
	if err != nil {
		return err
	}

	log.Infof("request: %+v", request)

	return nil
}
