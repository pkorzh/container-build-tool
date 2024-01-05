package archive

import (
	"fmt"
	"io"
	"os"

	"github.com/pkorzh/container-build-tool/internal/archive"
	"github.com/pkorzh/container-build-tool/internal/oci/internal"
	ocilayout "github.com/pkorzh/container-build-tool/internal/oci/layout"
	"github.com/pkorzh/container-build-tool/internal/tmpdir"
	"github.com/pkorzh/container-build-tool/internal/types"

	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type ociArchiveImageReader struct {
	ref                  ociArchiveRef
	ociLayoutImageReader types.ImageReader
	tmpDir               internal.TmpDirOCIRef
}

func (a ociArchiveImageReader) Close() error {
	defer a.tmpDir.DeleteTmpDir()
	return a.ociLayoutImageReader.Close()
}

func (a ociArchiveImageReader) GetBlob(d digest.Digest) (io.ReadCloser, error) {
	return a.ociLayoutImageReader.GetBlob(d)
}

func (a ociArchiveImageReader) GetManifest() (*imgspecv1.Manifest, error) {
	return a.ociLayoutImageReader.GetManifest()
}

func (a ociArchiveImageReader) GetImage() (*imgspecv1.Image, error) {
	return a.ociLayoutImageReader.GetImage()
}

func newImageReader(ref ociArchiveRef) (types.ImageReader, error) {
	tmpDirRef, err := untarIntoTmpDir(ref)
	if err != nil {
		return nil, err
	}

	ociLayoutImageReader, err := tmpDirRef.OCILayoutRef.NewImageReader()
	if err != nil {
		if err := tmpDirRef.DeleteTmpDir(); err != nil {
			return nil, err
		}
		return nil, err
	}

	return &ociArchiveImageReader{
		ref:                  ref,
		ociLayoutImageReader: ociLayoutImageReader,
		tmpDir:               tmpDirRef,
	}, nil
}

func untarIntoTmpDir(ref ociArchiveRef) (internal.TmpDirOCIRef, error) {
	src := ref.resolvedFile

	tmpdir, err := tmpdir.MkTmpDir("oci-archive")
	if err != nil {
		return internal.TmpDirOCIRef{}, fmt.Errorf("create tmp dir: %w", err)
	}

	arch, err := os.Open(src)
	if err != nil {
		return internal.TmpDirOCIRef{}, err
	}
	defer arch.Close()

	ociLayoutRef, err := ocilayout.NewReference(tmpdir, ref.image, ref.imageTag)
	if err != nil {
		return internal.TmpDirOCIRef{}, err
	}

	tempDirRef := internal.TmpDirOCIRef{TmpDir: tmpdir, OCILayoutRef: ociLayoutRef}

	err = archive.Untar(arch, tempDirRef.TmpDir)
	if err != nil {
		if err := tempDirRef.DeleteTmpDir(); err != nil {
			return internal.TmpDirOCIRef{}, fmt.Errorf("deleting tmp dir: %w", err)
		}
		return internal.TmpDirOCIRef{}, fmt.Errorf("untar")
	}

	return tempDirRef, nil
}
