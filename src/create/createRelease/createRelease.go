package createRelease

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
