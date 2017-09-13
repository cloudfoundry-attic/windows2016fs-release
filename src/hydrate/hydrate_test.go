package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/archiver/extractor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Hydrate", func() {
	extractTarball := func(path string) string {
		tmpDir, err := ioutil.TempDir("", "hydrated")
		Expect(err).NotTo(HaveOccurred())
		err = extractor.NewTgz().Extract(path, tmpDir)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
		return tmpDir
	}

	var outputTarball string

	BeforeEach(func() {
		tmpFile, err := ioutil.TempFile("", "rootfs")
		Expect(err).NotTo(HaveOccurred())
		outputTarball = tmpFile.Name()
		command := exec.Command(hydrateBin, "-output", outputTarball)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		Expect(os.Remove(outputTarball)).To(Succeed())
	})

	Context("Downloading the docker manifest", func() {
		It("is a valid manifest", func() {
			rootfs := extractTarball(outputTarball)
			defer os.RemoveAll(rootfs)

			manifestFile := filepath.Join(rootfs, "manifest.json")
			Expect(manifestFile).To(BeAnExistingFile())

			var dockerManifest struct {
				Name string
			}
			content, err := ioutil.ReadFile(manifestFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(json.Unmarshal(content, &dockerManifest)).To(Succeed())
			Expect(dockerManifest.Name).To(Equal("cloudfoundry/windows2016fs"))
		})
	})
})
