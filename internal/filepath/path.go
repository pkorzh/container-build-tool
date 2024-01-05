package filepath

import (
	"path/filepath"
)

func ResolvePath(path string) (string, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.Clean(resolved), nil
}
