package layout

import (
	"fmt"
	"path/filepath"

	_ "crypto/sha256"
	_ "crypto/sha512"

	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/opencontainers/go-digest"
	ctbfilepath "github.com/pkorzh/container-build-tool/internal/filepath"
	"github.com/pkorzh/container-build-tool/internal/json"
	"github.com/pkorzh/container-build-tool/internal/oci/internal"
	"github.com/pkorzh/container-build-tool/internal/types"
)

type ociLayoutRef struct {
	dir         string
	resolvedDir string
	image       string
	imageTag    string
}

func (ref ociLayoutRef) NewImageReader() (types.ImageReader, error) {
	return newImageReader(ref)
}

func (ref ociLayoutRef) NewImageWriter() (types.ImageWriter, error) {
	return newImageWriter(ref)
}

func (ref ociLayoutRef) ImageName() string {
	if ref.image == "" {
		dir := filepath.Base(ref.resolvedDir)
		return filepath.Base(dir)
	} else {
		return ref.image
	}
}

func (ref ociLayoutRef) indexPath() string {
	return filepath.Join(ref.resolvedDir, "index.json")
}

func (ref ociLayoutRef) ociLayoutPath() string {
	return filepath.Join(ref.dir, "oci-layout")
}

func (ref ociLayoutRef) blobPath(d digest.Digest) (string, error) {
	if err := d.Validate(); err != nil {
		return "", fmt.Errorf("unexpected digest reference %s: %w", d, err)
	}
	var blobDir = filepath.Join(ref.dir, "blobs")
	return filepath.Join(blobDir, d.Algorithm().String(), d.Hex()), nil
}

func (ref ociLayoutRef) index() (*imgspecv1.Index, error) {
	return json.ParseJSON[imgspecv1.Index](ref.indexPath())
}

func (ref ociLayoutRef) manifestDescriptor() (imgspecv1.Descriptor, error) {
	imageIndex, err := ref.index()
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	var annotationRefName = ""
	if ref.image != "" && ref.imageTag != "" {
		annotationRefName = fmt.Sprintf("%s:%s", ref.image, ref.imageTag)
	}

	if annotationRefName == "" {
		if len(imageIndex.Manifests) > 1 {
			return imgspecv1.Descriptor{}, fmt.Errorf("multiple images found in index, specify image name")
		}
		return imageIndex.Manifests[0], nil
	} else {
		for _, manifest := range imageIndex.Manifests {
			if v, ok := manifest.Annotations[imgspecv1.AnnotationRefName]; ok && v == annotationRefName {
				return manifest, nil
			}
		}
	}

	return imgspecv1.Descriptor{}, fmt.Errorf("image %s not found in index", annotationRefName)
}

func ParseReference(ref string) (types.ImageRef, error) {
	file, image, tag := internal.ExtractFileImageTag(ref)
	return NewReference(file, image, tag)
}

func NewReference(file, image, imageTag string) (types.ImageRef, error) {
	resolved, err := ctbfilepath.ResolvePath(file)
	if err != nil {
		return nil, err
	}
	return &ociLayoutRef{
		dir:         file,
		resolvedDir: resolved,
		image:       image,
		imageTag:    imageTag,
	}, nil
}
