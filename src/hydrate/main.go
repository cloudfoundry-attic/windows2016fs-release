package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/archiver/compressor"
)

const (
	ImageName   = "cloudfoundry/windows2016fs"
	ImageRef    = "latest"
	TokenURL    = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"
	ManifestURL = "https://registry.hub.docker.com/v2/%s/manifests/%s"
	BlobURL     = "https://registry.hub.docker.com/v2/%s/blobs/%s"
)

func main() {
	tempDir, err := ioutil.TempDir("", "hydrate")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	token, err := getToken(ImageName)
	if err != nil {
		panic(err)
	}

	manifest, err := getManifest(ImageName, ImageRef, token)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "manifest.json"), manifest, 0644)
	if err != nil {
		panic(err)
	}

	err = compressor.NewTgz().Compress(tempDir+"/", os.Args[2])
	if err != nil {
		panic(err)
	}
}

func getToken(imageName string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(TokenURL, ImageName))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var token struct {
		Token string
	}

	if err := json.Unmarshal(body, &token); err != nil {
		return "", err
	}

	return token.Token, nil
}

func getManifest(imageName, imageRef, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(ManifestURL, ImageName, ImageRef), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
