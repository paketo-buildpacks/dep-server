package utils_test

import (
	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils"
	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils/utilsfakes"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	spec.Run(t, "Utilities", testUtils, spec.Report(report.Terminal{}), spec.Parallel())
}

func testUtils(t *testing.T, context spec.G, it spec.S) {
	var (
		err                     error
		Expect                  = NewWithT(t).Expect
		phpUtils                utils.PHPExtensionsUtils
		fakeChecksummer         *dependencyfakes.FakeChecksummer
		fakeFileSystem          *dependencyfakes.FakeFileSystem
		fakeGithubClient        *dependencyfakes.FakeGithubClient
		fakeDepFactoryWebClient *dependencyfakes.FakeWebClient
		fakePHPExtensionsClient *utilsfakes.FakePHPExtensionsWebClient
		depFactory              dependency.DepFactory
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeDepFactoryWebClient = &dependencyfakes.FakeWebClient{}
		fakePHPExtensionsClient = &utilsfakes.FakePHPExtensionsWebClient{}
		depFactory = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeDepFactoryWebClient)
		phpUtils = utils.NewPHPExtensionsUtils(depFactory, fakePHPExtensionsClient)
	})

	context("When given a folder containing PHP extension metadata files", func() {
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

				PHPExtensionData, err := phpUtils.ParseYML(YMLFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(PHPExtensionData).To(Equal(utils.PHPExtMetadataFile{
					NativeModules: []utils.PHPExtension{
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
					Extensions: []utils.PHPExtension{
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
			it.Before(func() {
				//fakePHPExtensionsClient.DownloadExtensionsSourceReturns()

			})

			it("returns the latest upstream version of a given extension", func() {

				depVersion, err := phpUtils.GetLatestUpstreamVersion("apcu")
				Expect(err).NotTo(HaveOccurred())

				Expect(depVersion).To(Equal(utils.ExtensionVersion{
					Name:    "apcu",
					Version: "some-version",
					MD5:     "some-md5",
				}))
			})
		})

		context("GenerateJSONPayload", func() {})
	})
}
