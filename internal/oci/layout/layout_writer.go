package layout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/opencontainers/go-digest"
	imgspec "github.com/opencontainers/image-spec/specs-go"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkorzh/container-build-tool/internal/types"
)

type ociLayoutImageWriter struct {
	ref   ociLayoutRef
	index *imgspecv1.Index
}

func (a ociLayoutImageWriter) Close() error {
	return nil
}

func (a ociLayoutImageWriter) Save() error {
	layoutBytes, err := json.Marshal(imgspecv1.ImageLayout{
		Version: imgspecv1.ImageLayoutVersion,
	})
	if err != nil {
		return err
	}

	if err := os.WriteFile(a.ref.ociLayoutPath(), layoutBytes, 0644); err != nil {
		return err
	}

	indexJSON, err := json.Marshal(a.index)
	if err != nil {
		return err
	}
	err = os.WriteFile(a.ref.indexPath(), indexJSON, 0644)
	if err != nil {
		return err
	}

	blobsPath := filepath.Join(a.ref.dir, "blobs")
	if _, err = os.Stat(blobsPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(blobsPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (a ociLayoutImageWriter) PutImageBlob(i imgspecv1.Image, m *imgspecv1.Manifest) (imgspecv1.Descriptor, error) {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	descriptor, err := a.PutBlob(bytes.NewReader(jsonBytes), types.PutBlobOptions{
		MediaType: imgspecv1.MediaTypeImageConfig,
	})
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	m.Config = descriptor

	return descriptor, nil
}

func (a ociLayoutImageWriter) PutManifestBlob(m imgspecv1.Manifest) (imgspecv1.Descriptor, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	descriptor, err := a.PutBlob(bytes.NewReader(jsonBytes), types.PutBlobOptions{
		MediaType: imgspecv1.MediaTypeImageManifest,
	})
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	if descriptor.Annotations != nil && descriptor.Annotations[imgspecv1.AnnotationRefName] != "" {
		for i, m := range a.index.Manifests {
			if m.Annotations[imgspecv1.AnnotationRefName] == descriptor.Annotations[imgspecv1.AnnotationRefName] {
				delete(a.index.Manifests[i].Annotations, imgspecv1.AnnotationRefName)
				break
			}
		}
	}

	a.index.Manifests = append(a.index.Manifests, descriptor)

	return descriptor, nil
}

func (a ociLayoutImageWriter) PutBlob(blob io.Reader, options types.PutBlobOptions) (imgspecv1.Descriptor, error) {
	var tmpFileClosed bool

	tmpFile, err := os.CreateTemp(a.ref.dir, "oci-layout-blob-")
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	defer func() {
		if !tmpFileClosed {
			tmpFile.Close()
		}
		os.Remove(tmpFile.Name())
	}()

	digester := digest.Canonical.Digester()
	multiWriter := io.MultiWriter(tmpFile, digester.Hash())

	size, err := io.Copy(multiWriter, blob)
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	if err := tmpFile.Sync(); err != nil {
		return imgspecv1.Descriptor{}, err
	}

	blobDigest := digester.Digest()

	blobPath, err := a.ref.blobPath(blobDigest)
	if err != nil {
		return imgspecv1.Descriptor{}, err
	}

	blogPathDir := filepath.Dir(blobPath)
	if _, err := os.Stat(blogPathDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(blogPathDir, 0755); err != nil {
			return imgspecv1.Descriptor{}, fmt.Errorf("creating blob dir: %w", err)
		}
	}

	tmpFile.Close()
	tmpFileClosed = true

	if err := os.Rename(tmpFile.Name(), blobPath); err != nil {
		return imgspecv1.Descriptor{}, err
	}

	return imgspecv1.Descriptor{
		Digest:      blobDigest,
		Size:        size,
		MediaType:   options.MediaType,
		Annotations: options.Annotations,
	}, nil
}

func (a ociLayoutImageWriter) GetBlob(d digest.Digest) (io.ReadCloser, error) {
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

func newImageWriter(ref ociLayoutRef) (types.ImageWriter, error) {
	var index *imgspecv1.Index

	if _, err := os.Stat(ref.indexPath()); err != nil && os.IsNotExist(err) {
		index = &imgspecv1.Index{
			Versioned: imgspec.Versioned{
				SchemaVersion: 2,
			},
		}
	} else {
		index, err = ref.index()
		if err != nil {
			return nil, err
		}
	}

	return &ociLayoutImageWriter{
		ref:   ref,
		index: index,
	}, nil
}
