package main

import (
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	var fromCmd = &cobra.Command{
		Use:           "from",
		Short:         "Create a working container based on an image.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			return handleFromCmd(c, args)
		},
		Example: `cbt from nginx:latest
cbt from node:18`,
	}

	rootCmd.AddCommand(fromCmd)
}

func handleFromCmd(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("an image name must be specified")
	}

	if len(args) > 1 {
		return errors.New("too many arguments specified")
	}

	return nil
}
