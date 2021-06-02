package main_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"

	. "github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint"
)

func TestUtils(t *testing.T) {
	spec.Run(t, "Utilities", testUtils, spec.Report(report.Terminal{}), spec.Parallel())
}

func testUtils(t *testing.T, context spec.G, it spec.S) {
	var (
		err    error
		Expect = NewWithT(t).Expect
	)

	context("When given a folder cotaining PHP extension metadata files", func() {
		var folderPath string
		var YMLFile string

		context("ParseYML", func() {
			it.Before(func() {
				folderPath, err = ioutil.TempDir("", "php-extensions")
				Expect(err).NotTo(HaveOccurred())
				YMLFile = filepath.Join(folderPath, "test.yml")
				err = ioutil.WriteFile(YMLFile, []byte(`native_modules:
  - name: hiredis
    version: 1.0.0
    md5: hiredisMD5
    klass: HiredisRecipe
  - name: librdkafka
    version: 1.5.2
    md5: librdkafkaMD5
    klass: LibRdKafkaRecipe
extensions:
  - name: apcu
    version: 5.1.19
    md5: apcuMD5
    klass: PeclRecipe
  - name: cassandra
    version: 1.3.2
    md5: cassandraMD5
    klass: PeclRecipe
`), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("Should parse the PHP extension metadata YML to an Object", func() {

				PHPExtensionData, err := ParseYML(YMLFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(PHPExtensionData).To(Equal(PHPExtMetadataFile{
					NativeModules: []PHPExtension{
						{
							Name:    "hiredis",
							Version: "1.0.0",
							MD5:     "hiredisMD5",
							Klass:   "HiredisRecipe",
						},
						{
							Name:    "librdkafka",
							Version: "1.5.2",
							MD5:     "librdkafkaMD5",
							Klass:   "LibRdKafkaRecipe",
						},
					},
					Extensions: []PHPExtension{
						{
							Name:    "apcu",
							Version: "5.1.19",
							MD5:     "apcuMD5",
							Klass:   "PeclRecipe",
						},
						{
							Name:    "cassandra",
							Version: "1.3.2",
							MD5:     "cassandraMD5",
							Klass:   "PeclRecipe",
						},
					},
				}))
			})

			it.After(func() {
				Expect(os.RemoveAll(folderPath)).To(Succeed())
			})
		})

		context("GetLatestUpstreamVersion", func() {
			it("returns the latest upstream version of a given extension", func() {

				depVersion, err := GetLatestUpstreamVersion("apcu")
				Expect(err).NotTo(HaveOccurred())

				Expect(depVersion).To(Equal(ExtensionVersion{
					Name:    "apcu",
					Version: "some-version",
					MD5:     "some-md5",
				}))
			})
		})

		context("GenerateJSONPayload", func() {})
	})
}
