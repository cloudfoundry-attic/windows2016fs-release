package main

import (
	"bufio"
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
)

func main() {
	rootfstgz := os.Args[1]
	outputDir := os.Args[2]

	layerTempDir, err := ioutil.TempDir("", "hcslayers")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("extracting %s to %s...\n", rootfstgz, layerTempDir)
	ex := extractor.NewTgz()
	if err := ex.Extract(rootfstgz, layerTempDir); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	//fmt.Printf("Done. deleting %s\n", rootfstgz)
	//if err := os.Remove(rootfstgz); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}

	type manifest struct {
		Layers []string `json:"Layers"`
	}

	manifestData, err := ioutil.ReadFile(filepath.Join(layerTempDir, "manifest.json"))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var m []manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	layerTarFiles := m[0].Layers
	var layerIds []string
	for _, f := range layerTarFiles {
		layerIds = append(layerIds, filepath.Dir(f))
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	parentLayers := []string{}
	for _, layer := range layerIds {
		if err := extractLayer(layerTempDir, outputDir, layer, parentLayers); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		parentLayers = append([]string{layer}, parentLayers...)
	}

	//if err := os.RemoveAll(layerTempDir); err != nil {
	//	fmt.Println(err.Error())
	//	os.Exit(1)
	//}
}

func extractLayer(layerTarDir, outputDir, layerId string, parentLayerIds []string) error {
	tarFile := filepath.Join(layerTarDir, layerId, "layer.tar")
	layerDir := filepath.Join(outputDir, layerId)
	fmt.Printf("Extracting layerID %s\n\tfrom: %s\n\tto: %s\n", layerId, tarFile, layerDir)

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

	fmt.Printf("parentLayerPaths: %+v\n", parentLayerPaths)

	layerWriter, err := hcsshim.NewLayerWriter(di, layerId, parentLayerPaths)
	if err != nil {
		return err
	}

	tf, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer tf.Close()

	tr := tar.NewReader(tf)

	hdr, err := tr.Next()
	buf := bufio.NewWriter(nil)

	for err == nil {
		base := path.Base(hdr.Name)
		if strings.HasPrefix(base, ".wh.") {
			name := path.Join(path.Dir(hdr.Name), base[len(".wh."):])
			err = layerWriter.Remove(filepath.FromSlash(name))
			if err != nil {
				return err
			}
			hdr, err = tr.Next()
		} else if hdr.Typeflag == tar.TypeLink {
			err = layerWriter.AddLink(filepath.FromSlash(hdr.Name), filepath.FromSlash(hdr.Linkname))
			if err != nil {
				return err
			}
			hdr, err = tr.Next()
		} else {
			var (
				name     string
				fileInfo *winio.FileBasicInfo
			)
			name, _, fileInfo, err = backuptar.FileInfoFromHeader(hdr)
			if err != nil {
				return err
			}
			err = layerWriter.Add(filepath.FromSlash(name), fileInfo)
			if err != nil {
				return err
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
