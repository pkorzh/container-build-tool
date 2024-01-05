package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/mattn/go-shellwords"
	"github.com/pkorzh/container-build-tool/internal/builder"
	"github.com/spf13/cobra"
)

type configFlags struct {
	workingDir string
	user       string
	cmd        string
	entrypoint string
	ports      []string
	os         string
	arch       string
}

func init() {
	var opts configFlags

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Config OCI image",
		RunE: func(c *cobra.Command, args []string) error {
			return handleConfigCmd(c, args, opts)
		},
		Args: cobra.ExactArgs(1),
	}

	flags := configCmd.Flags()
	flags.StringVar(&opts.workingDir, "workingdir", "", "Working directory")
	flags.StringVar(&opts.user, "user", "", "User")
	flags.StringVar(&opts.cmd, "cmd", "", "Command")
	flags.StringVar(&opts.entrypoint, "entrypoint", "", "Entrypoint")
	flags.StringSliceVar(&opts.ports, "ports", []string{}, "Ports")
	flags.StringVar(&opts.os, "os", runtime.GOOS, "OS")
	flags.StringVar(&opts.arch, "arch", runtime.GOARCH, "Architecture")

	rootCmd.AddCommand(configCmd)
}

func handleConfigCmd(c *cobra.Command, args []string, opts configFlags) error {
	builder, err := builder.Open(args[0])
	if err != nil {
		return err
	}

	if c.Flag("workingdir").Changed {
		builder.SetWorkingDir(opts.workingDir)
	}

	if c.Flag("user").Changed {
		builder.SetUser(opts.user)
	}

	if c.Flag("cmd").Changed {
		cmdSpec, err := shellwords.Parse(opts.cmd)
		if err != nil {
			return fmt.Errorf("parsing cmd %q: %w", opts.cmd, err)
		}
		builder.SetCmd(cmdSpec)
	}

	if c.Flag("entrypoint").Changed {
		entrypointJson := []string{}
		err := json.Unmarshal([]byte(opts.entrypoint), &entrypointJson)
		if err == nil {
			builder.SetEntrypoint(entrypointJson)
		} else {
			entrypointArr := make([]string, 3)
			entrypointArr[0] = "/bin/sh"
			entrypointArr[1] = "-c"
			entrypointArr[2] = opts.entrypoint
			builder.SetEntrypoint(entrypointArr)
		}
	}

	if c.Flag("ports").Changed {
		builder.SetPorts(opts.ports)
	}

	if c.Flag("os").Changed {
		builder.SetOS(opts.os)
	}

	if c.Flag("arch").Changed {
		builder.SetArch(opts.arch)
	}

	err = builder.Save()
	if err != nil {
		return err
	}

	return nil
}
