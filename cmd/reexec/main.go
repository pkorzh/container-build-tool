package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "ns":
		ns()
	default:
		panic("specify run command")
	}
}
func run() {
	fmt.Printf("Running %v\n", os.Args[2:])
	cmd := exec.Command("/proc/self/exe", append([]string{"ns"},
		os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	cmd.Run()
}
func ns() {
	fmt.Printf("Running in new UTS namespace %v\n", os.Args[2:])

	syscall.Sethostname([]byte("inside-container"))
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}
