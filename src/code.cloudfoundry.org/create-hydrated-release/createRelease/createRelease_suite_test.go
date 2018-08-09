package createRelease_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCreateRelease(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CreateRelease Suite")
}
