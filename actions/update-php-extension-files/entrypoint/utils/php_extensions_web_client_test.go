package utils_test

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPHPExtensionsWebClient(t *testing.T) {
	spec.Run(t, "PHPExtensionsWebClient", testPHPExtensionsWebClient, spec.Report(report.Terminal{}), spec.Parallel())
}

func testPHPExtensionsWebClient(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect                 = NewWithT(t).Expect
		client                 *httptest.Server
		testDir                string
		phpExtensionsWebClient utils.WebClient
	)

	const (
		fileContents = "some-contents"
	)

	context("When using the testPHPExtensionsWebClient", func() {
		it.Before(func() {
			var err error
			testDir, err = ioutil.TempDir("", "external-dependency-resource-web-client")
			Expect(err).NotTo(HaveOccurred())

			client = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/file":
					_, _ = fmt.Fprint(w, fileContents)
				case "/headers":
					_, _ = fmt.Fprint(w, r.Header)
				case "/body":
					defer r.Body.Close()
					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())
					_, _ = fmt.Fprint(w, string(body))
				case "/500":
					w.WriteHeader(500)
					_, _ = fmt.Fprint(w, "some-server-error")
				}
			}))

			phpExtensionsWebClient = utils.NewPHPExtensionsWebClient()
		})

		context("DownloadExtensionsSource downloads the PHP extension source", func() {
			it("downloads the file", func() {
				outputPath := filepath.Join(testDir, "some-file.txt")

				err := phpExtensionsWebClient.DownloadExtensionsSource(client.URL+"/file", outputPath)
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(outputPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(fileContents))
			})
		})

		context("DownloadExtensionsSource response is not 200", func() {
			it("returns an error", func() {
				err := phpExtensionsWebClient.DownloadExtensionsSource(client.URL+"/500", "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("got unsuccessful response: status code: 500, body: some-server-error")))
			})
		})

		it.After(func() {
			_ = os.RemoveAll(testDir)
			client.Close()
		})
	})
}
