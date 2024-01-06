package layer

import (
	"io"

	"github.com/opencontainers/go-digest"
	"github.com/pkorzh/container-build-tool/internal/archive"
)

type LayerInfo struct {
	CompressedDigest digest.Digest
	CompressedSize   int64

	UncompressedDigest digest.Digest

	MediaType string
}

type writeCounter struct {
	Writer io.Writer
	Count  int64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n, err := wc.Writer.Write(p)
	wc.Count += int64(n)
	return n, err
}

func GetLayerInfo(layer io.Reader) (LayerInfo, error) {
	compressedDigester := digest.Canonical.Digester()
	counter := &writeCounter{Writer: compressedDigester.Hash()}
	teeWriter := io.TeeReader(layer, counter)

	arch, _, err := archive.DecompressStream(teeWriter)
	if err != nil {
		return LayerInfo{}, err
	}

	decompressedDigester := digest.Canonical.Digester()
	decompressedHash := decompressedDigester.Hash()

	_, err = io.Copy(decompressedHash, arch)
	if err != nil {
		return LayerInfo{}, err
	}

	return LayerInfo{
		CompressedDigest:   compressedDigester.Digest(),
		CompressedSize:     counter.Count,
		UncompressedDigest: decompressedDigester.Digest(),
	}, nil
}
