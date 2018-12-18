package main_test

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func cp(src string, dst string) {
	session, err := gexec.Start(exec.Command("powershell", "-Command", "Copy-Item", "-Recurse", "-Force", src, dst),
		GinkgoWriter,
		GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	Eventually(session, 5*time.Minute).Should(gexec.Exit(0))
}
