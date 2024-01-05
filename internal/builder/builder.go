package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pkorzh/container-build-tool/internal/image"
	"github.com/pkorzh/container-build-tool/internal/workdir"

	imgspec "github.com/opencontainers/image-spec/specs-go"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type BuilderOptions struct {
	FromImage string
}

type BuildOptions struct {
	Target string
	Layers []string
}

type Builder struct {
	FromImage   string              `json:"fromImage"`
	WorkDirID   string              `json:"workDirId"`
	OCIImage    *imgspecv1.Image    `json:"ociImage"`
	OCIManifest *imgspecv1.Manifest `json:"ociManifest"`
}

func (b *Builder) Save() error {
	obj, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshalling builder: %w", err)
	}

	workDir, err := workdir.GetWorkingContainerDir(b.WorkDirID)
	if err != nil {
		return fmt.Errorf("getting workdir: %w", err)
	}

	err = os.WriteFile(workDir+"/builder.json", obj, 0600)
	if err != nil {
		return fmt.Errorf("writing builder: %w", err)
	}

	return nil
}

func New(options BuilderOptions) (*Builder, error) {
	imageRef, err := image.ParseReference(options.FromImage)
	if err != nil {
		return nil, fmt.Errorf("parsing image reference: %w", err)
	}

	workDirId := imageRef.ImageName()
	workDir, err := workdir.NewWorkingContainerDir(workDirId)
	if err != nil {
		return nil, fmt.Errorf("creating workdir: %w", err)
	}

	imageReader, err := imageRef.NewImageReader()
	if err != nil {
		if err := os.RemoveAll(workDir); err != nil {
			return nil, fmt.Errorf("removing workdir: %w", err)
		}
		return nil, fmt.Errorf("creating image reader: %w", err)
	}
	defer imageReader.Close()

	now := time.Now().UTC()

	fromImage, err := imageReader.GetImage()
	if err != nil {
		if err := os.RemoveAll(workDir); err != nil {
			return nil, fmt.Errorf("removing workdir: %w", err)
		}
		return nil, fmt.Errorf("getting image: %w", err)
	}

	return &Builder{
		FromImage: options.FromImage,
		WorkDirID: workDirId,
		OCIImage: &imgspecv1.Image{
			Created:      &now,
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			RootFS:       imgspecv1.RootFS{Type: "layers"},
			Config:       fromImage.Config,
		},
		OCIManifest: &imgspecv1.Manifest{
			Versioned: imgspec.Versioned{
				SchemaVersion: 2,
			},
			MediaType: imgspecv1.MediaTypeImageManifest,
		},
	}, nil
}

func Open(workDirID string) (*Builder, error) {
	workDir, err := workdir.GetWorkingContainerDir(workDirID)
	if err != nil {
		return nil, fmt.Errorf("getting workdir: %w", err)
	}

	obj, err := os.ReadFile(workDir + "/builder.json")
	if err != nil {
		return nil, fmt.Errorf("reading builder: %w", err)
	}

	var builder Builder
	err = json.Unmarshal(obj, &builder)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling builder: %w", err)
	}

	return &builder, nil
}
