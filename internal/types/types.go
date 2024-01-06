package types

import (
	"io"

	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ImageReader interface {
	Close() error
	GetBlob(digest.Digest) (io.ReadCloser, error)
	GetManifest() (*imgspecv1.Manifest, error)
	GetImage() (*imgspecv1.Image, error)
}

type PutBlobOptions struct {
	Annotations map[string]string
	MediaType   string
}

type ImageWriter interface {
	Close() error
	Save() error
	PutManifestBlob(imgspecv1.Manifest) (imgspecv1.Descriptor, error)
	PutImageBlob(imgspecv1.Image, *imgspecv1.Manifest) (imgspecv1.Descriptor, error)
	PutBlob(io.Reader, PutBlobOptions) (imgspecv1.Descriptor, error)
	GetBlob(digest.Digest) (io.ReadCloser, error)
}

type ImageRef interface {
	NewImageReader() (ImageReader, error)
	NewImageWriter() (ImageWriter, error)
	ImageName() string
}
