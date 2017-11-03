package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/windows2016fs/hydrator"
	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-utils/logger"
)

func CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir string) {
	tagData, err := ioutil.ReadFile(imageTagPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	imageTag := string(tagData)

	h := hydrator.New(outputDir, imageName, imageTag)
	if err := h.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	versionData, err := ioutil.ReadFile(versionDataPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	releaseVersion := cmd.VersionArg{}
	if err := releaseVersion.UnmarshalFlag(string(versionData)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	l := logger.NewLogger(logger.LevelInfo)
	u := ui.NewConfUI(l)
	defer u.Flush()
	deps := cmd.NewBasicDeps(u, l)

	createReleaseOpts := &cmd.CreateReleaseOpts{
		Directory: cmd.DirOrCWDArg{Path: releaseDir},
		Version:   releaseVersion,
	}

	if tarballPath != "" {
		expanded, err := filepath.Abs(tarballPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		createReleaseOpts.Tarball = cmd.FileArg{FS: deps.FS, ExpandedPath: expanded}
	}

	createReleaseCommand := cmd.NewCmd(cmd.BoshOpts{}, createReleaseOpts, deps)
	if err := createReleaseCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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

	CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir)
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
