package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/archiver/extractor"

	winio "github.com/Microsoft/go-winio"
	"github.com/Microsoft/go-winio/archive/tar"
	"github.com/Microsoft/go-winio/backuptar"
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

	fmt.Fprintf(os.Stderr, "extracting %s to %s...\n", rootfstgz, layerTempDir)
	ex := extractor.NewTgz()
	if err := ex.Extract(rootfstgz, layerTempDir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	manifestData, err := ioutil.ReadFile(filepath.Join(layerTempDir, "manifest.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var m v1.Manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	parentLayers := []string{}
	for _, layer := range m.Layers {
		layerId := layer.Digest.Encoded()

		if err := extractLayer(layerTempDir, outputDir, layerId, parentLayers); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		parentLayers = append([]string{layerId}, parentLayers...)
	}

	if err := os.RemoveAll(layerTempDir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	toplayer := parentLayers[0]

	fmt.Printf(filepath.Join(outputDir, toplayer))
}

func extractLayer(layerTarDir, outputDir, layerId string, parentLayerIds []string) error {
	gzipFile := filepath.Join(layerTarDir, layerId)

	layerDir := filepath.Join(outputDir, layerId)
	fmt.Fprintf(os.Stderr, "Extracting layerID %s\n\tfrom: %s\n\tto: %s\n", layerId, gzipFile, layerDir)

	if err := os.MkdirAll(layerDir, 0755); err != nil {
		return err
	}

	if err := winio.EnableProcessPrivileges([]string{winio.SeBackupPrivilege, winio.SeRestorePrivilege}); err != nil {
		return err
	}
	defer winio.DisableProcessPrivileges([]string{winio.SeBackupPrivilege, winio.SeRestorePrivilege})

	di := hcsshim.DriverInfo{HomeDir: outputDir, Flavour: 1}

	parentLayerPaths := []string{}
	for _, id := range parentLayerIds {
		parentLayerPaths = append(parentLayerPaths, filepath.Join(outputDir, id))
	}

	fmt.Fprintf(os.Stderr, "parentLayerPaths: %+v\n", parentLayerPaths)

	if len(parentLayerPaths) > 0 {
		data, err := json.Marshal(parentLayerPaths)
		if err != nil {
			return fmt.Errorf("Failed to marshal layerchain.json: %s", err.Error())
		}

		if err := ioutil.WriteFile(filepath.Join(outputDir, layerId, "layerchain.json"), data, 0644); err != nil {
			return fmt.Errorf("Failed to write layerchain.json: %s", err.Error())
		}
	}

	layerWriter, err := hcsshim.NewLayerWriter(di, layerId, parentLayerPaths)
	if err != nil {
		return fmt.Errorf("Failed to create new layer writer: %s", err.Error())
	}

	gf, err := os.Open(gzipFile)
	if err != nil {
		return err
	}
	defer gf.Close()

	gr, err := gzip.NewReader(gf)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	hdr, err := tr.Next()
	buf := bufio.NewWriter(nil)

	for err == nil {
		base := path.Base(hdr.Name)
		if strings.HasPrefix(base, ".wh.") {
			name := path.Join(path.Dir(hdr.Name), base[len(".wh."):])
			err = layerWriter.Remove(filepath.FromSlash(name))
			if err != nil {
				return fmt.Errorf("Failed to remove: %s", err.Error())
			}
			hdr, err = tr.Next()
		} else if hdr.Typeflag == tar.TypeLink {
			err = layerWriter.AddLink(filepath.FromSlash(hdr.Name), filepath.FromSlash(hdr.Linkname))
			if err != nil {
				return fmt.Errorf("Failed to add link: %s", err.Error())
			}
			hdr, err = tr.Next()
		} else {
			var (
				name     string
				fileInfo *winio.FileBasicInfo
			)
			name, _, fileInfo, err = backuptar.FileInfoFromHeader(hdr)
			if err != nil {
				return fmt.Errorf("Failed to get file info: %s", err.Error())
			}
			err = layerWriter.Add(filepath.FromSlash(name), fileInfo)
			if err != nil {
				return fmt.Errorf("Failed to add layer: %s", err.Error())
			}
			buf.Reset(layerWriter)

			hdr, err = backuptar.WriteBackupStreamFromTarFile(buf, tr, hdr)
			ferr := buf.Flush()
			if ferr != nil {
				err = ferr
			}
		}
	}

	if err != io.EOF {
		return err
	}

	return layerWriter.Close()
}
