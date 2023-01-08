package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&adapter, "adapter", "a", "postgres", "The database adapter to use")

	viper.BindPFlag("adapter", rootCmd.PersistentFlags().Lookup("adapter"))
}

var (
	adapter string
)

var rootCmd = &cobra.Command{
	Use:   "terramigrate",
	Short: "Database migration tool",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
