package layer_test

import (
	"io/ioutil"
	"oci-image/layer"
	"os"
	"path/filepath"

	"github.com/Microsoft/hcsshim"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Layer", func() {
	var (
		lm        *layer.Manager
		outputDir string
	)

	BeforeEach(func() {
		var err error
		outputDir, err = ioutil.TempDir("", "oci-image.layer_test")
		Expect(err).NotTo(HaveOccurred())

		lm = layer.NewManager(hcsshim.DriverInfo{HomeDir: outputDir})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(outputDir)).To(Succeed())
	})

	Describe("State", func() {
		const layerId = "layer1"

		Context("the layer directory does not exist", func() {
			It("returns NotExist", func() {
				state, err := lm.State(layerId)
				Expect(err).NotTo(HaveOccurred())
				Expect(state).To(Equal(layer.State(layer.NotExist)))
			})
		})

		Context("the layer directory exists", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(outputDir, layerId), 0755)).To(Succeed())
			})

			Context("and does not contain a .complete file", func() {
				It("returns Incomplete", func() {
					state, err := lm.State(layerId)
					Expect(err).NotTo(HaveOccurred())
					Expect(state).To(Equal(layer.State(layer.Incomplete)))
				})
			})

			Context("and contains a .complete file with the wrong id", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(outputDir, layerId, ".complete"), []byte("abcde"), 0644))
				})

				It("returns Incomplete", func() {
					state, err := lm.State(layerId)
					Expect(err).NotTo(HaveOccurred())
					Expect(state).To(Equal(layer.State(layer.Incomplete)))
				})
			})

			Context("and contains a .complete file with the matching id", func() {
				BeforeEach(func() {
					Expect(ioutil.WriteFile(filepath.Join(outputDir, layerId, ".complete"), []byte(layerId), 0644))
				})

				It("returns Valid", func() {
					state, err := lm.State(layerId)
					Expect(err).NotTo(HaveOccurred())
					Expect(state).To(Equal(layer.State(layer.Valid)))
				})
			})
		})
	})
})
