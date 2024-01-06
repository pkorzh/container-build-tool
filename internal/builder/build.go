package builder

import (
	"fmt"
	"path/filepath"

	"github.com/pkorzh/container-build-tool/internal/archive"
	"github.com/pkorzh/container-build-tool/internal/image"
	"github.com/pkorzh/container-build-tool/internal/layer"
	"github.com/pkorzh/container-build-tool/internal/types"
	"github.com/pkorzh/container-build-tool/internal/workdir"

	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func (b *Builder) Build(options BuildOptions) error {
	dstImageRef, err := image.ParseReference(options.Target)
	if err != nil {
		return fmt.Errorf("parsing image reference: %w", err)
	}

	dstImageWriter, err := dstImageRef.NewImageWriter()
	if err != nil {
		return fmt.Errorf("creating image reader: %w", err)
	}
	defer dstImageWriter.Close()

	srcImageRef, err := image.ParseReference(b.FromImage)
	if err != nil {
		return fmt.Errorf("parsing image reference: %w", err)
	}

	srcImageReader, err := srcImageRef.NewImageReader()
	if err != nil {
		return fmt.Errorf("creating image reader: %w", err)
	}
	defer srcImageReader.Close()

	rootFSLayers, err := b.copyRootFsBlobs(dstImageWriter, srcImageReader)
	if err != nil {
		return fmt.Errorf("copying rootfs blobs: %w", err)
	}

	usersLayers, err := b.copyUsersFsBlobs(options.Layers, dstImageWriter)
	if err != nil {
		return fmt.Errorf("copying users blobs: %w", err)
	}

	b.addLayers(rootFSLayers)
	b.addLayers(usersLayers)

	dstImageWriter.PutImageBlob(*b.OCIImage, b.OCIManifest)
	dstImageWriter.PutManifestBlob(*b.OCIManifest)

	err = dstImageWriter.Save()
	if err != nil {
		return fmt.Errorf("saving image: %w", err)
	}

	return nil
}

func (b *Builder) addLayers(layers []layer.LayerInfo) {
	for _, layer := range layers {
		b.OCIImage.RootFS.DiffIDs = append(b.OCIImage.RootFS.DiffIDs, layer.UncompressedDigest)

		b.OCIManifest.Layers = append(b.OCIManifest.Layers, imgspecv1.Descriptor{
			MediaType: layer.MediaType,
			Digest:    layer.CompressedDigest,
			Size:      layer.CompressedSize,
		})
	}
}

func (b *Builder) copyRootFsBlobs(writer types.ImageWriter, reader types.ImageReader) ([]layer.LayerInfo, error) {
	srcManifest, err := reader.GetManifest()
	if err != nil {
		return nil, err
	}

	srcImage, err := reader.GetImage()
	if err != nil {
		return nil, err
	}

	layerInfos := make([]layer.LayerInfo, 0, len(srcManifest.Layers))

	for i, layerDescriptor := range srcManifest.Layers {
		blobReader, err := reader.GetBlob(layerDescriptor.Digest)
		if err != nil {
			return nil, fmt.Errorf("getting blob: %w", err)
		}
		defer blobReader.Close()

		blobDescriptor, err := writer.PutBlob(blobReader, types.PutBlobOptions{
			MediaType: layerDescriptor.MediaType,
		})
		if err != nil {
			return nil, fmt.Errorf("put blog: %w", err)
		}

		if layerDescriptor.Digest != blobDescriptor.Digest {
			return nil, fmt.Errorf("digest mismatch: %s != %s", layerDescriptor.Digest, blobDescriptor.Digest)
		}

		layerInfos = append(layerInfos, layer.LayerInfo{
			MediaType:          blobDescriptor.MediaType,
			CompressedDigest:   blobDescriptor.Digest,
			CompressedSize:     blobDescriptor.Size,
			UncompressedDigest: srcImage.RootFS.DiffIDs[i],
		})
	}

	return layerInfos, nil
}

func (b *Builder) copyUsersFsBlobs(layerDirNames []string, writer types.ImageWriter) ([]layer.LayerInfo, error) {
	var layerInfos []layer.LayerInfo

	workdir, err := workdir.GetWorkingContainerDir(b.WorkDirID)
	if err != nil {
		return nil, fmt.Errorf("getting workdir: %w", err)
	}

	for _, layerDirName := range layerDirNames {
		layerDir := filepath.Join(workdir, "layers", layerDirName)

		layerInfo, err := func() (layer.LayerInfo, error) {
			arch, err := archive.Tar(layerDir, archive.Gzip)
			if err != nil {
				return layer.LayerInfo{}, fmt.Errorf("archiving layer: %w", err)
			}
			defer arch.Close()

			descriptor, err := writer.PutBlob(arch, types.PutBlobOptions{
				MediaType: imgspecv1.MediaTypeImageLayerGzip,
			})
			if err != nil {
				return layer.LayerInfo{}, fmt.Errorf("putting blob: %w", err)
			}

			blob, err := writer.GetBlob(descriptor.Digest)
			if err != nil {
				return layer.LayerInfo{}, fmt.Errorf("getting blob: %w", err)
			}
			defer blob.Close()

			layerInfo, err := layer.GetLayerInfo(blob)
			if err != nil {
				return layer.LayerInfo{}, fmt.Errorf("getting layer info: %w", err)
			}

			layerInfo.MediaType = descriptor.MediaType

			return layerInfo, nil
		}()

		if err != nil {
			return nil, fmt.Errorf("archiving: %w", err)
		}

		layerInfos = append(layerInfos, layerInfo)
	}

	return layerInfos, nil
}
