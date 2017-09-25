package main

import (
	"flag"
	"fmt"
	"hydrate/compress"
	"hydrate/hydrator"
	"hydrate/registry"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	outDir, imageName, imageTag := parseFlags()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fatalErr("Could not create output directory")
	}

	if imageName == "" {
		fatalErr("No image name provided")
	}

	nameParts := strings.Split(imageName, "/")
	if len(nameParts) != 2 {
		fatalErr("Invalid image name")
	}
	outFile := filepath.Join(outDir, fmt.Sprintf("%s-%s.tgz", nameParts[1], imageTag))

	tempDir, err := ioutil.TempDir("", "hydrate")
	if err != nil {
		fatalErr(fmt.Sprintf("Could not create tmp dir: %s", tempDir))
	}
	defer os.RemoveAll(tempDir)

	r := registry.New("https://auth.docker.io", "https://registry.hub.docker.com", imageName, imageTag)
	c := compress.New()
	h := hydrator.New(tempDir, outFile, r, c, os.Stdout)

	if err := h.Run(); err != nil {
		fatalErr(err.Error())
	}
	fmt.Println("Done.")
	os.Exit(0)
}

func parseFlags() (string, string, string) {
	outDir := flag.String("outputDir", os.TempDir(), "Output directory for downloaded image")
	imageName := flag.String("image", "", "Name of the image to download")
	imageTag := flag.String("tag", "latest", "Image tag to download")
	flag.Parse()

	return *outDir, *imageName, *imageTag
}

func fatalErr(msg string) {
	fmt.Fprintln(os.Stderr, "ERROR: "+msg)
	os.Exit(1)
}
