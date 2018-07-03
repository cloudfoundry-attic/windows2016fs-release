// +build !windows

package main_test

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func cp(src string, dst string) {
	session, err := gexec.Start(
		exec.Command("cp", "-a", src, dst),
		GinkgoWriter,
		GinkgoWriter,
	)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(0))
}
