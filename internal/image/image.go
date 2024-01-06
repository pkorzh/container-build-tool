package image

import (
	"fmt"
	"strings"

	"github.com/pkorzh/container-build-tool/internal/oci/archive"
	"github.com/pkorzh/container-build-tool/internal/oci/layout"
	"github.com/pkorzh/container-build-tool/internal/types"
)

func ParseReference(ref string) (types.ImageRef, error) {
	source, fromImage, found := strings.Cut(ref, ":")
	if !found {
		return nil, fmt.Errorf("invalid image reference: %s", ref)
	}
	switch source {
	case "oci-archive":
		return archive.ParseReference(fromImage)
	case "oci-layout":
		return layout.ParseReference(fromImage)
	default:
		return nil, fmt.Errorf("invalid image reference: %s", ref)
	}
}
