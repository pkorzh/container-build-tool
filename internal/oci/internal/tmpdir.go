package internal

import (
	"os"

	"github.com/pkorzh/container-build-tool/internal/types"
)

type TmpDirOCIRef struct {
	TmpDir       string
	OCILayoutRef types.ImageRef
}

func (t TmpDirOCIRef) DeleteTmpDir() error {
	return os.RemoveAll(t.TmpDir)
}
