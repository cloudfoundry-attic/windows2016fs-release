package createRelease

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/hydrator/hydrator"
	"github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-utils/logger"

	yaml "gopkg.in/yaml.v2"
)

type ReleaseCreator struct{}

func (rc ReleaseCreator) CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath string) error {
	tagData, err := ioutil.ReadFile(imageTagPath)
	if err != nil {
		return err
	}
	imageTag := string(tagData)

	releaseName, _ := rc.GetReleaseName(releaseDir)
	h := hydrator.New(log.New(os.Stdout, "", 0), filepath.Join(releaseDir, "blobs", releaseName), imageName, imageTag, false)
	if err := h.Run(); err != nil {
		return err
	}

	versionData, err := ioutil.ReadFile(versionDataPath)
	if err != nil {
		return err
	}

	releaseVersion := cmd.VersionArg{}
	if err := releaseVersion.UnmarshalFlag(string(versionData)); err != nil {
		return err
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
			return err
		}

		createReleaseOpts.Tarball = cmd.FileArg{FS: deps.FS, ExpandedPath: expanded}
	}

	// bosh create-release adds ~7GB of temp files that should be cleaned up
	tmpDir, err := ioutil.TempDir("", "winfs-create-release")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	os.Setenv("HOME", tmpDir)

	createReleaseCommand := cmd.NewCmd(cmd.BoshOpts{}, createReleaseOpts, deps)
	if err := createReleaseCommand.Execute(); err != nil {
		return err
	}

	return nil
}

func (rc ReleaseCreator) GetReleaseName(releaseDir string) (string, error) {
	fileContents, err := ioutil.ReadFile(filepath.Join(releaseDir, "config", "final.yml"))
	if err != nil {
		return "", err
	}

	yamlObj := make(map[string]interface{})
	err = yaml.Unmarshal(fileContents, yamlObj)
	if err != nil {
		return "", err
	}

	name, bExists := yamlObj["name"].(string)
	if !bExists {
		// Yaml doesn't contain the name
		return "", err
	}

	return name, nil
}
