package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	yaml "gopkg.in/yaml.v2"
)

var _ = Describe("Create", func() {
	var (
		releaseDir string
		version    string
		session    *gexec.Session
	)

	type releaseYaml struct {
		Version           string `yaml:"version"`
		UncommitedChanges bool   `yaml:"uncommitted_changes"`
	}

	BeforeEach(func() {
		var err error
		releaseDir, err = ioutil.TempDir("", "create.test-release")
		Expect(err).NotTo(HaveOccurred())

		srcReleaseDir := filepath.Join("..", "..")
		copyReleaseDir(srcReleaseDir, releaseDir)
		checkoutDirectory(releaseDir)

		versionData, err := ioutil.ReadFile(filepath.Join(releaseDir, "VERSION"))
		Expect(err).NotTo(HaveOccurred())
		version = string(versionData)

		cmd := exec.Command(createBin, "--releaseDir", releaseDir)

		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(releaseDir)).To(Succeed())
	})

	It("creates the release", func() {
		Eventually(session, 30*time.Minute).Should(gexec.Exit(0))

		data, err := ioutil.ReadFile(filepath.Join(releaseDir, "dev_releases", "windows2016fs", fmt.Sprintf("windows2016fs-%s.yml", version)))
		Expect(err).NotTo(HaveOccurred())

		var release releaseYaml
		Expect(yaml.Unmarshal(data, &release)).To(Succeed())

		Expect(release.Version).To(Equal(version))
		Expect(release.UncommitedChanges).To(BeFalse())
	})
})

func copyReleaseDir(src, dst string) {
	pathsToCopy := []string{
		"config",
		".git",
		".gitignore",
		".gitmodules",
		"jobs",
		"packages",
		"src",
		"VERSION",
	}
	for _, path := range pathsToCopy {
		cp(filepath.Join(src, path), dst)
	}
}

func checkoutDirectory(dir string) {
	cmds := []*exec.Cmd{
		exec.Command("git", "config", "core.filemode", "false"),
		exec.Command("git", "config", "user.email", "garden-windows-eng@pivotal.io"),
		exec.Command("git", "config", "user.name", "Garden Windows CI"),
		exec.Command("git", "submodule", "foreach", "--recursive", "git", "config", "core.filemode", "false"),
		exec.Command("git", "add", "."),
		exec.Command("git", "commit", "--allow-empty", "-m", "WIP - test commit"),
	}

	for _, cmd := range cmds {
		cmd.Dir = dir
		gitSession, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(gitSession, 2*time.Minute).Should(gexec.Exit(0))
	}
}
