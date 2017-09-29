package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"oci-image/image"
	"oci-image/layer"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/archiver/extractor"

	"github.com/Microsoft/hcsshim"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

func main() {
	rootfstgz := os.Args[1]
	outputDir := os.Args[2]

	layerTempDir, err := ioutil.TempDir("", "hcslayers")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := extractor.NewTgz().Extract(rootfstgz, layerTempDir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer os.RemoveAll(layerTempDir)

	manifestData, err := ioutil.ReadFile(filepath.Join(layerTempDir, "manifest.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	lm := layer.NewManager(hcsshim.DriverInfo{HomeDir: outputDir, Flavour: 1})
	im := image.NewManager(layerTempDir, outputDir, manifest, lm, os.Stderr)

	topLayer, err := im.Extract()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Printf(topLayer)
}
