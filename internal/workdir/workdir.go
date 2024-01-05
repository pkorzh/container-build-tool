package workdir

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func baseDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(user.HomeDir, ".cbt"), nil
}

func ensureWorkdirBaseExists() (string, error) {
	baseDir, err := baseDir()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		return "", err
	}

	return baseDir, nil
}

func NewWorkingContainerDir(name string) (string, error) {
	baseDir, err := ensureWorkdirBaseExists()
	if err != nil {
		return "", err
	}

	workingContainerDir := filepath.Join(baseDir, name)
	_, err = os.Stat(workingContainerDir)
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("working container directory already exists")
	}

	if err := os.MkdirAll(workingContainerDir, 0755); err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Join(workingContainerDir, "layers"), 0755); err != nil {
		return "", err
	}

	return workingContainerDir, nil
}

func GetWorkingContainerDir(name string) (string, error) {
	baseDir, err := ensureWorkdirBaseExists()
	if err != nil {
		return "", err
	}

	workingContainerDir := filepath.Join(baseDir, name)
	_, err = os.Stat(workingContainerDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("working container directory does not exist")
	}

	return workingContainerDir, nil
}
