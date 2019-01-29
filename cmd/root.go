package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:           "go-ard",
	Short:         "Architecture Decision Records manager",
	Long:          "Architecture Decision Records manager",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.PersistentFlags().Bool("debug", false, "debug mode (default: false)")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", false)
}