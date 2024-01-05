package layout

import (
	"io"
	"os"

	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkorzh/container-build-tool/internal/json"
	"github.com/pkorzh/container-build-tool/internal/types"
)

type ociLayoutImageReader struct {
	ref        ociLayoutRef
	index      *imgspecv1.Index
	descriptor imgspecv1.Descriptor
}

func (a ociLayoutImageReader) Close() error {
	return nil
}

func (a ociLayoutImageReader) Index() *imgspecv1.Index {
	return a.index
}

func (a ociLayoutImageReader) GetBlob(d digest.Digest) (io.ReadCloser, error) {
	blobPath, err := a.ref.blobPath(d)
	if err != nil {
		return nil, err
	}

	contents, err := os.Open(blobPath)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (a ociLayoutImageReader) GetManifest() (*imgspecv1.Manifest, error) {
	manifestPath, err := a.ref.blobPath(a.descriptor.Digest)
	if err != nil {
		return nil, err
	}

	manifest, err := json.ParseJSON[imgspecv1.Manifest](manifestPath)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func (a ociLayoutImageReader) GetImage() (*imgspecv1.Image, error) {
	manifest, err := a.GetManifest()
	if err != nil {
		return nil, err
	}

	imagePath, err := a.ref.blobPath(manifest.Config.Digest)
	if err != nil {
		return nil, err
	}

	image, err := json.ParseJSON[imgspecv1.Image](imagePath)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func newImageReader(ref ociLayoutRef) (types.ImageReader, error) {
	index, err := ref.index()
	if err != nil {
		return nil, err
	}

	menifestDescriptor, err := ref.manifestDescriptor()
	if err != nil {
		return nil, err
	}

	return &ociLayoutImageReader{
		ref:        ref,
		index:      index,
		descriptor: menifestDescriptor,
	}, nil
}
