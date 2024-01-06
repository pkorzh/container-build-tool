package archive

import (
	"io"
	"os"

	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkorzh/container-build-tool/internal/archive"
	"github.com/pkorzh/container-build-tool/internal/oci/internal"
	ocilayout "github.com/pkorzh/container-build-tool/internal/oci/layout"
	"github.com/pkorzh/container-build-tool/internal/tmpdir"
	"github.com/pkorzh/container-build-tool/internal/types"
)

type ociArchiveImageWriter struct {
	ref                  ociArchiveRef
	ociLayoutImageWriter types.ImageWriter
	tmpDir               internal.TmpDirOCIRef
}

func (a ociArchiveImageWriter) Close() error {
	defer a.tmpDir.DeleteTmpDir()
	return a.ociLayoutImageWriter.Close()
}

func (a ociArchiveImageWriter) Save() error {
	err := a.ociLayoutImageWriter.Save()
	if err != nil {
		return err
	}

	src := a.tmpDir.TmpDir
	dst := a.ref.resolvedFile

	file, err := os.Create(dst)
	if err != nil {
		return nil
	}
	defer file.Close()

	reader, err := archive.Tar(src, archive.Gzip)
	if err != nil {
		return err
	}
	defer reader.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return err
	}

	return nil
}

func (a ociArchiveImageWriter) PutImageBlob(i imgspecv1.Image, m *imgspecv1.Manifest) (imgspecv1.Descriptor, error) {
	return a.ociLayoutImageWriter.PutImageBlob(i, m)
}

func (a ociArchiveImageWriter) PutManifestBlob(m imgspecv1.Manifest) (imgspecv1.Descriptor, error) {
	return a.ociLayoutImageWriter.PutManifestBlob(m)
}

func (a ociArchiveImageWriter) PutBlob(r io.Reader, options types.PutBlobOptions) (imgspecv1.Descriptor, error) {
	return a.ociLayoutImageWriter.PutBlob(r, options)
}

func (a ociArchiveImageWriter) GetBlob(digest digest.Digest) (io.ReadCloser, error) {
	return a.ociLayoutImageWriter.GetBlob(digest)
}

func newImageWriter(ref ociArchiveRef) (types.ImageWriter, error) {
	tmpdir, err := tmpdir.MkTmpDir("oci-archive")
	if err != nil {
		return nil, err
	}

	ociLayoutRef, err := ocilayout.NewReference(tmpdir, ref.image, ref.imageTag)
	if err != nil {
		if err := os.RemoveAll(tmpdir); err != nil {
			return nil, err
		}
		return nil, err
	}

	imageWriter, err := ociLayoutRef.NewImageWriter()
	if err != nil {
		if err := os.RemoveAll(tmpdir); err != nil {
			return nil, err
		}
		return nil, err
	}

	return &ociArchiveImageWriter{
		ref:                  ref,
		ociLayoutImageWriter: imageWriter,
		tmpDir: internal.TmpDirOCIRef{
			TmpDir: tmpdir,
		},
	}, nil
}
