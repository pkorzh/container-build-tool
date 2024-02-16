package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

//The c code is posted at the end of the page.

/*
#include <stdlib.h>
void exec_ps();
void create_argv(int len);
void set_argv(int pos, char *arg);
*/
import "C"

var rootCmd = &cobra.Command{
	Use:          "minicr",
	Long:         "Mini containers runtime",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		//We re-execute self, effectively running a bootstrap process.
		// For alternative implementation please see docker's [reexec](https://github.com/moby/moby/blob/master/pkg/reexec/reexec.go) package.
		reexec := exec.Command("/proc/self/exe", "bootstrap", args[0])
		reexec.Stderr = os.Stderr
		reexec.Stdin = os.Stdin
		reexec.Stdout = os.Stdout

		reexec.SysProcAttr = &unix.SysProcAttr{
			Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWNS | unix.CLONE_NEWPID,
		}

		return reexec.Run()
	},
}

//This is the bootstrap process. The Command is hidden so that end users don't see it.
var bootstrapCmd = &cobra.Command{
	Use:    "bootstrap",
	Long:   "Configure namespaces",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		unix.Sethostname([]byte("inside-container"))

		//Since parent's mount list can be `shared` we need to make it private in our namespace.
		if err := unix.Mount("", "/", "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
			return err
		}

		pivotRoot(args[0])

		//Mount `/proc` so that we can use `ps aux`.
		if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
			return err
		}

		args = []string{"/bin/bash"}

		C.create_argv(C.int(len(args)))
		for i, arg := range args {
			cArg := C.CString(arg)
			C.set_argv(C.int(i), cArg)
			defer C.free(unsafe.Pointer(cArg))
		}

		C.exec_ps()

		return nil
	},
}

func pivotRoot(root string) error {
	// We need this to satisfy restriction: `new_root` and `put_old` must **not** be on the same filesystem as the current root.
	if err := unix.Mount(root, root, "bind", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself: %v", err)
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	//`syscall.PivotRoot` call changes the root mount in the mount namespace of the calling process.
	// It moves the root mount to the directory `.pivot_root` and makes `root` the new root mount.
	// Afterwards we can unmount `.pivot_root`, aka the old root.
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	return os.Remove(pivotDir)
}

func main() {
	//`LockOSThread` wires the calling goroutine to its current operating system thread.
	// The calling goroutine will always execute in that thread, and no other goroutine will execute in it.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rootCmd.AddCommand(bootstrapCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
