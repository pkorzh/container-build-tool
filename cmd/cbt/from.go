package main

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/pkorzh/container-build-tool/internal/builder"
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
		Example: `cbt from oci-archive:/tmp/centos.tar
cbt from oci-layout:/tmp/centos:latest
cbt from oci-layout:/tmp/nodejs:nodejs:latest`,
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

	builderOptions := builder.BuilderOptions{
		FromImage: args[0],
	}

	builder, err := builder.New(builderOptions)
	if err != nil {
		return err
	}

	err = builder.Save()
	if err != nil {
		return err
	}

	return nil
}
