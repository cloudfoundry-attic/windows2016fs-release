package main_test

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestHydrate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hydrate Suite")
}

var hydrateBin string

var _ = BeforeSuite(func() {
	var err error
	hydrateBin, err = gexec.Build("hydrate")
	Expect(err).NotTo(HaveOccurred())
	rand.Seed(time.Now().UnixNano())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
