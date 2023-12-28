package main

import (
	"github.com/spf13/cobra"
)

func init() {
	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build an image from a working container.",
		RunE: func(c *cobra.Command, args []string) error {
			return handleBuildCmd(c, args)
		},
		Args: cobra.MinimumNArgs(2),
	}

	rootCmd.AddCommand(buildCmd)
}

func handleBuildCmd(c *cobra.Command, args []string) error {
	return nil
}
