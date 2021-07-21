package licenses_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/licenses"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBundlerLicenseCase(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		mockServer       *httptest.Server
		licenseRetriever licenses.LicenseRetriever
	)

	it.Before(func() {
		var err error
		licenseRetriever = licenses.NewLicenseRetriever()
		// Set up tar files
		buffer := bytes.NewBuffer(nil)
		tw := tar.NewWriter(buffer)

		licenseFile := "LICENSE"
		licenseContent, err := os.ReadFile(filepath.Join("testdata", "LICENSE.md"))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(licenseContent))})).To(Succeed())
		_, err = tw.Write(licenseContent)
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())

		outerBuffer := bytes.NewBuffer(nil)
		outerTw := tar.NewWriter(outerBuffer)

		Expect(outerTw.WriteHeader(&tar.Header{Name: "data.tar.gz", Mode: 0755, Size: int64(buffer.Len())})).To(Succeed())
		_, err = outerTw.Write(buffer.Bytes())
		Expect(err).NotTo(HaveOccurred())

		Expect(outerTw.Close()).To(Succeed())

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)

				return
			}

			switch req.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)

			case "/bundler-source-url.gem":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, outerBuffer.String())

			case "/non-tar-file-outer-artifact":
				w.WriteHeader(http.StatusOK)
				// return a flac header, which is an unrecognized mime-type
				fmt.Fprint(w, bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22")))

			case "/non-tar-file-inner-artifact":
				outerBuffer := bytes.NewBuffer(nil)
				outerTw := tar.NewWriter(outerBuffer)

				flacHeader := []byte("\x66\x4C\x61\x43\x00\x00\x00\x22")

				Expect(outerTw.WriteHeader(&tar.Header{Name: "data.tar.gz", Mode: 0755, Size: int64(len(flacHeader))})).To(Succeed())
				_, err = outerTw.Write(flacHeader)
				Expect(err).NotTo(HaveOccurred())

				Expect(outerTw.Close()).To(Succeed())

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, outerBuffer.String())

			case "/no-license.tgz":
				// Set up tar files
				buffer := bytes.NewBuffer(nil)
				tw := tar.NewWriter(buffer)

				Expect(tw.WriteHeader(&tar.Header{Name: "./", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
				_, err = tw.Write(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tw.Close()).To(Succeed())

				outerBuffer := bytes.NewBuffer(nil)
				outerTw := tar.NewWriter(outerBuffer)

				Expect(outerTw.WriteHeader(&tar.Header{Name: "data.tar.gz", Mode: 0755, Size: int64(buffer.Len())})).To(Succeed())
				_, err = outerTw.Write(buffer.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(outerTw.Close()).To(Succeed())

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, outerBuffer.String())

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))
	})

	it.After(func() {
		mockServer.Close()
	})

	context("given a bundler dependency URL to get the license for", func() {
		it("gets the artifact and retrieves the license from it", func() {
			licenses, err := licenseRetriever.LookupLicenses("bundler", fmt.Sprintf("%s/bundler-source-url.gem", mockServer.URL))
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(Equal([]string{"MIT", "MIT-0"}))
		})
	})

	context("the artifact does not contain a license", func() {
		it("returns an empty slice of licenses and no error", func() {
			licenses, err := licenseRetriever.LookupLicenses("bundler", fmt.Sprintf("%s/no-license.tgz", mockServer.URL))
			Expect(err).ToNot(HaveOccurred())
			Expect(licenses).To(Equal([]string{}))
		})
	})

	context("failure cases", func() {
		context("the outer artifact cannot be decompressed", func() {
			it("returns an error", func() {
				_, err := licenseRetriever.LookupLicenses("bundler", fmt.Sprintf("%s/non-tar-file-outer-artifact", mockServer.URL))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("failed to decompress source file")))
			})
		})

		context("the inner artifact cannot be decompressed", func() {
			it("returns an error", func() {
				_, err := licenseRetriever.LookupLicenses("bundler", fmt.Sprintf("%s/non-tar-file-inner-artifact", mockServer.URL))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("failed to decompress inner source file")))
			})
		})
	})
}
