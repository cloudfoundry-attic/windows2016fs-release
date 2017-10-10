package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/windows2016fs/hydrator"
	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-utils/logger"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("bad args")
		os.Exit(1)
	}

	releaseDir := os.Args[1]
	imageName := "cloudfoundry/windows2016fs"
	tagData, err := ioutil.ReadFile(filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "IMAGE_TAG"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	imageTag := string(tagData)

	outputDir := filepath.Join(releaseDir, "blobs", "windows2016fs")

	h := hydrator.New(outputDir, imageName, imageTag)
	if err := h.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	versionData, err := ioutil.ReadFile(filepath.Join(releaseDir, "VERSION"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	releaseVersion := cmd.VersionArg{}
	if err := releaseVersion.UnmarshalFlag(string(versionData)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	createReleaseOpts := &cmd.CreateReleaseOpts{
		Directory: cmd.DirOrCWDArg{Path: releaseDir},
		Version:   releaseVersion,
	}

	l := logger.NewLogger(logger.LevelInfo)
	u := ui.NewConfUI(l)
	defer u.Flush()
	deps := cmd.NewBasicDeps(u, l)
	createReleaseCommand := cmd.NewCmd(cmd.BoshOpts{}, createReleaseOpts, deps)
	if err := createReleaseCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
