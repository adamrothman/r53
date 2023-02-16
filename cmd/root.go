package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Flags
	hostedZone string

	rootCmd = &cobra.Command{
		Use:   "r53",
		Short: "A tool that facilitates interactions with Route 53",
	}
)

func init() {
	f := rootCmd.PersistentFlags()

	f.StringVarP(&hostedZone, "zone", "z", "", "hosted zone ID")
	rootCmd.MarkPersistentFlagRequired("zone")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
