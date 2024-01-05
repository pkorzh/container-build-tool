package tmpdir

import (
	"os"
	"runtime"
)

func getTmpDir() string {
	if runtime.GOOS == "windows" {
		return os.TempDir()
	} else {
		return "/var/tmp"
	}
}

func MkTmpDir(name string) (string, error) {
	return os.MkdirTemp(getTmpDir(), "cbt-"+name)
}
