package layer

import "github.com/pkorzh/container-build-tool/internal/types"

func ReadFromImageReader(reader types.ImageReader) ([]LayerInfo, error) {
	manifest, err := reader.GetManifest()
	if err != nil {
		return nil, err
	}

	layerInfos := make([]LayerInfo, 0, len(manifest.Layers))

	for _, layer := range manifest.Layers {
		blobReader, err := reader.GetBlob(layer.Digest)
		if err != nil {
			return nil, err
		}

		layerInfo, err := GetLayerInfo(blobReader)
		blobReader.Close()

		if err != nil {
			return nil, err
		}

		layerInfos = append(layerInfos, layerInfo)
	}

	return layerInfos, nil
}
