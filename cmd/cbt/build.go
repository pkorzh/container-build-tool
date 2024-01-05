package main

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/pkorzh/container-build-tool/internal/builder"
)

type buildFlags struct {
	layers []string
}

func init() {
	var opts buildFlags
	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build an image from a working container.",
		RunE: func(c *cobra.Command, args []string) error {
			return handleBuildCmd(c, args, opts)
		},
		Args: cobra.ExactArgs(2),
	}

	flags := buildCmd.Flags()
	flags.StringSliceVar(&opts.layers, "layers", []string{}, "Layers to add to the image")

	rootCmd.AddCommand(buildCmd)
}

func handleBuildCmd(c *cobra.Command, args []string, opts buildFlags) error {
	b, err := builder.Open(args[0])
	if err != nil {
		return err
	}

	if !c.Flag("layers").Changed {
		return errors.New("layers must be specified")
	}

	buildOptions := builder.BuildOptions{
		Target: args[1],
		Layers: opts.layers,
	}

	err = b.Build(buildOptions)
	if err != nil {
		return err
	}

	return nil
}
