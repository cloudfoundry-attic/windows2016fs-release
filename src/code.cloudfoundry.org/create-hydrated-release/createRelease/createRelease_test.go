package createRelease_test

import (
	"path/filepath"

	"code.cloudfoundry.org/create-hydrated-release/createRelease"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseCreator", func() {
	Describe("GetReleaseName", func() {
		It("returns the release name from the release directory", func() {
			rc := &createRelease.ReleaseCreator{}
			releaseDir := filepath.Join("..", "..", "..", "..")
			releaseName, _ := rc.GetReleaseName(releaseDir)
			Expect(releaseName).To(HavePrefix("windows"))
		})
	})
})
