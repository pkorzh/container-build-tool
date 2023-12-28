package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type runFlags struct {
	lowerLayers []string
	upperLayer  string
}

func init() {
	var opts runFlags
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Runs specified command in an isolated environment aka container.",
		RunE: func(c *cobra.Command, args []string) error {
			return handleRunCmd(c, args, opts)
		},
		Args: cobra.MinimumNArgs(1),
		Example: `CONTAINER=$(cbt from ubuntu:latest)
cbt run $CONTAINER /bin/bash
cbt run $CONTAINER -l root -u deps -- /bin/sh -c "echo hello world"`,
	}

	flags := runCmd.Flags()
	flags.SetInterspersed(false)
	flags.StringSliceVarP(&opts.lowerLayers, "lower-layers", "l", []string{"root"}, "Specify lower layers to use")
	flags.StringVarP(&opts.upperLayer, "upper-layer", "u", "", "Specify upper layer to use")

	rootCmd.AddCommand(runCmd)
}

func handleRunCmd(c *cobra.Command, args []string, opts runFlags) error {
	fmt.Printf("Running %v\n", args)
	fmt.Printf("Opts: %v\n", opts)
	return nil
}
