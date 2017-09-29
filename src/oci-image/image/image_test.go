package image_test

import (
	"encoding/json"
	"io/ioutil"
	"oci-image/image"
	"oci-image/image/imagefakes"
	"oci-image/layer"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

var _ = Describe("Image", func() {
	Describe("Extract", func() {
		var (
			srcDir   string
			destDir  string
			tempDir  string
			manifest v1.Manifest
			lm       *imagefakes.FakeLayerManager
			e        *image.Extractor
		)

		const (
			layer1 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			layer2 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
			layer3 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "oci-image.image")
			Expect(err).NotTo(HaveOccurred())

			layers := []v1.Descriptor{
				{Digest: digest.NewDigestFromEncoded("sha256", layer1)},
				{Digest: digest.NewDigestFromEncoded("sha256", layer2)},
				{Digest: digest.NewDigestFromEncoded("sha256", layer3)},
			}

			manifest = v1.Manifest{
				Layers: layers,
			}

			srcDir = filepath.Join(tempDir, "src")
			destDir = filepath.Join(tempDir, "dest")
			lm = &imagefakes.FakeLayerManager{}

			e = image.NewExtractor(srcDir, destDir, manifest, lm, ioutil.Discard)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		It("extracts all the layers, returning the top", func() {
			topLayer, err := e.Extract()
			Expect(err).NotTo(HaveOccurred())
			Expect(topLayer).To(Equal(filepath.Join(destDir, layer3)))

			Expect(lm.DeleteCallCount()).To(Equal(0))

			Expect(lm.ExtractCallCount()).To(Equal(3))
			tgz, id, parentPaths := lm.ExtractArgsForCall(0)
			Expect(tgz).To(Equal(filepath.Join(srcDir, layer1)))
			Expect(id).To(Equal(layer1))
			Expect(parentPaths).To(Equal([]string{}))

			tgz, id, parentPaths = lm.ExtractArgsForCall(1)
			Expect(tgz).To(Equal(filepath.Join(srcDir, layer2)))
			Expect(id).To(Equal(layer2))
			Expect(parentPaths).To(Equal([]string{filepath.Join(destDir, layer1)}))

			tgz, id, parentPaths = lm.ExtractArgsForCall(2)
			Expect(tgz).To(Equal(filepath.Join(srcDir, layer3)))
			Expect(id).To(Equal(layer3))
			Expect(parentPaths).To(Equal([]string{filepath.Join(destDir, layer2), filepath.Join(destDir, layer1)}))
		})

		It("creates the proper destination directory for each layer", func() {
			lm.ExtractStub = func(_, id string, _ []string) error {
				Expect(filepath.Join(destDir, id)).To(BeADirectory())
				return nil
			}

			_, err := e.Extract()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("a layer has already been extracted", func() {
			BeforeEach(func() {
				lm.StateReturnsOnCall(1, layer.Valid, nil)
			})

			It("does not re-extract the existing layer", func() {
				_, err := e.Extract()
				Expect(err).NotTo(HaveOccurred())
				Expect(lm.DeleteCallCount()).To(Equal(0))

				Expect(lm.ExtractCallCount()).To(Equal(2))
				tgz, id, parentPaths := lm.ExtractArgsForCall(0)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer1)))
				Expect(id).To(Equal(layer1))
				Expect(parentPaths).To(Equal([]string{}))

				tgz, id, parentPaths = lm.ExtractArgsForCall(1)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer3)))
				Expect(id).To(Equal(layer3))
				Expect(parentPaths).To(Equal([]string{filepath.Join(destDir, layer2), filepath.Join(destDir, layer1)}))
			})
		})

		Context("there is an invalid layer", func() {
			BeforeEach(func() {
				lm.StateReturnsOnCall(1, layer.Incomplete, nil)
			})

			It("deletes the incomplete layer and re-extracts", func() {
				_, err := e.Extract()
				Expect(err).NotTo(HaveOccurred())

				Expect(lm.DeleteCallCount()).To(Equal(1))
				Expect(lm.DeleteArgsForCall(0)).To(Equal(layer2))

				Expect(lm.ExtractCallCount()).To(Equal(3))
				tgz, id, parentPaths := lm.ExtractArgsForCall(0)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer1)))
				Expect(id).To(Equal(layer1))
				Expect(parentPaths).To(Equal([]string{}))

				tgz, id, parentPaths = lm.ExtractArgsForCall(1)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer2)))
				Expect(id).To(Equal(layer2))
				Expect(parentPaths).To(Equal([]string{filepath.Join(destDir, layer1)}))

				tgz, id, parentPaths = lm.ExtractArgsForCall(2)
				Expect(tgz).To(Equal(filepath.Join(srcDir, layer3)))
				Expect(id).To(Equal(layer3))
				Expect(parentPaths).To(Equal([]string{filepath.Join(destDir, layer2), filepath.Join(destDir, layer1)}))
			})
		})
	})
})

func getLayerChain(dir string) []string {
	var layers []string
	data, err := ioutil.ReadFile(filepath.Join(dir, "layerchain.json"))
	Expect(err).NotTo(HaveOccurred())

	Expect(json.Unmarshal(data, &layers)).To(Succeed())
	return layers
}
