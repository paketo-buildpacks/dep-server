package main_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBundlerLicenseCase(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		mockServer *httptest.Server
	)

	it.Before(func() {
		var err error
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
			command := exec.Command(
				entrypoint,
				"--dependency-name", "bundler",
				"--url", fmt.Sprintf("%s/bundler-source-url.gem", mockServer.URL),
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("Getting the dependency artifact from %s/bundler-source-url.gem", mockServer.URL)))

			Expect(buffer).To(gbytes.Say("Decompressing the dependency artifact"))
			Expect(buffer).To(gbytes.Say("Scanning artifact for license file"))

			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::\[MIT MIT-0\]`))
			Expect(buffer).To(gbytes.Say("Licenses found!"))
		})
	})

	context("failure cases", func() {
		context("the outer artifact cannot be decompressed", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "bundler",
					"--url", fmt.Sprintf("%s/non-tar-file-outer-artifact", mockServer.URL),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`failed to decompress source file`))
			})
		})

		context("the inner artifact cannot be decompressed", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "bundler",
					"--url", fmt.Sprintf("%s/non-tar-file-inner-artifact", mockServer.URL),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`failed to decompress inner source file`))
			})
		})

		context("the bundler artifact does not contain a license", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "bundler",
					"--url", fmt.Sprintf("%s/no-license.tgz", mockServer.URL),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`failed to detect licenses: no license file was found`))
			})
		})
	})
}
