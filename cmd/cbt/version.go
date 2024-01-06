package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays version information.",
		RunE: func(c *cobra.Command, args []string) error {
			return handleVersionCmd()
		},
		Args:    cobra.NoArgs,
		Example: `buildah version`,
	}

	rootCmd.AddCommand(versionCmd)
}

func handleVersionCmd() error {
	fmt.Println("Runtime version: ", runtime.Version())
	fmt.Println("OS/Arch:         ", runtime.GOOS+"/"+runtime.GOARCH)
	return nil
}
