package image

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate counterfeiter . LayerManager
type LayerManager interface {
	Extract(string, string, []string) error
}

type Extractor struct {
	srcDir       string
	outputDir    string
	manifest     v1.Manifest
	layerManager LayerManager
	output       io.Writer
}

func NewExtractor(srcDir, outputDir string, manifest v1.Manifest, layerManager LayerManager, output io.Writer) *Extractor {
	return &Extractor{
		srcDir:       srcDir,
		outputDir:    outputDir,
		manifest:     manifest,
		layerManager: layerManager,
		output:       output,
	}
}

func (e *Extractor) Extract() (string, error) {
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return "", err
	}

	parentLayerPaths := []string{}
	for _, layer := range e.manifest.Layers {
		layerId := layer.Digest.Encoded()
		layerTgz := filepath.Join(e.srcDir, layerId)
		layerDir := filepath.Join(e.outputDir, layerId)

		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return "", err
		}

		fmt.Fprintf(e.output, "Extracting %s... ", layerId)
		if err := e.layerManager.Extract(layerTgz, layerId, parentLayerPaths); err != nil {
			return "", err
		}
		fmt.Fprintln(e.output, "Done.")

		if len(parentLayerPaths) > 0 {
			data, err := json.Marshal(parentLayerPaths)
			if err != nil {
				return "", fmt.Errorf("Failed to marshal layerchain.json: %s", err.Error())
			}

			if err := ioutil.WriteFile(filepath.Join(e.outputDir, layerId, "layerchain.json"), data, 0644); err != nil {
				return "", fmt.Errorf("Failed to write layerchain.json: %s", err.Error())
			}
		}

		parentLayerPaths = append([]string{layerDir}, parentLayerPaths...)
	}

	return parentLayerPaths[0], nil
}
