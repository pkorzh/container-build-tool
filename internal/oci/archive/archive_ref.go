package archive

import (
	"path/filepath"

	internalfilepath "github.com/pkorzh/container-build-tool/internal/filepath"
	"github.com/pkorzh/container-build-tool/internal/oci/internal"
	"github.com/pkorzh/container-build-tool/internal/types"
)

type ociArchiveRef struct {
	file         string
	resolvedFile string
	image        string
	imageTag     string
}

func (ref ociArchiveRef) NewImageReader() (types.ImageReader, error) {
	return newImageReader(ref)
}

func (ref ociArchiveRef) NewImageWriter() (types.ImageWriter, error) {
	return newImageWriter(ref)
}

func (ref ociArchiveRef) ImageName() string {
	if ref.image == "" {
		fileName := filepath.Base(ref.resolvedFile)
		fileExt := filepath.Ext(fileName)
		return fileName[:len(fileName)-len(fileExt)]
	} else {
		return ref.image
	}
}

func ParseReference(ref string) (types.ImageRef, error) {
	file, image, tag := internal.ExtractFileImageTag(ref)
	return NewReference(file, image, tag)
}

func NewReference(file, image, imageTag string) (types.ImageRef, error) {
	resolved, err := internalfilepath.ResolvePath(file)
	if err != nil {
		return nil, err
	}
	return ociArchiveRef{
		file:         file,
		resolvedFile: resolved,
		image:        image,
		imageTag:     imageTag,
	}, nil
}
