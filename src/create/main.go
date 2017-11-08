package main

import (
	"create/createRelease"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	releaseDir, tarballPath, err := parseArgs()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	imageName := "cloudfoundry/windows2016fs"

	imageTagPath := filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "IMAGE_TAG")

	outputDir := filepath.Join(releaseDir, "blobs", "windows2016fs")

	versionDataPath := filepath.Join(releaseDir, "VERSION")

	releaseCreator := new(createRelease.ReleaseCreator)
	err = releaseCreator.CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func parseArgs() (string, string, error) {
	var releaseDir, tarballPath string
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	flagSet.StringVar(&releaseDir, "releaseDir", "", "")
	flagSet.StringVar(&tarballPath, "tarball", "", "")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return "", "", err
	}

	if releaseDir == "" {
		return "", "", errors.New("missing required flag 'releaseDir'")
	}

	return releaseDir, tarballPath, nil
}
