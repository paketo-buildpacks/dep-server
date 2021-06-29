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

func testDefaultLicenseCase(t *testing.T, context spec.G, it spec.S) {
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

		Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
		_, err = tw.Write(nil)
		Expect(err).NotTo(HaveOccurred())

		licenseFile := filepath.Join("some-dir", "LICENSE")
		licenseContent, err := os.ReadFile(filepath.Join("testdata", "LICENSE"))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(licenseContent))})).To(Succeed())
		_, err = tw.Write(licenseContent)
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)

				return
			}

			switch req.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)

			case "/default-dependency-source-url.tgz":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, buffer.String())

			case "/dotnet-source-url.tgz":
				buffer := bytes.NewBuffer(nil)
				tw := tar.NewWriter(buffer)

				licenseContent, err := os.ReadFile(filepath.Join("testdata", "LICENSE"))
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.WriteHeader(&tar.Header{Name: "LICENSE", Mode: 0755, Size: int64(len(licenseContent))})).To(Succeed())
				_, err = tw.Write(licenseContent)
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.Close()).To(Succeed())
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, buffer.String())

			case "/bad-url":
				w.WriteHeader(http.StatusBadRequest)

			case "/non-tar-file-artifact":
				w.WriteHeader(http.StatusOK)
				// return a flac header, which is an unrecognized mime-type
				fmt.Fprint(w, bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22")))

			case "/no-license.tgz":
				buffer = bytes.NewBuffer(nil)
				tw = tar.NewWriter(buffer)

				Expect(tw.WriteHeader(&tar.Header{Name: "./", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
				_, err = tw.Write(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(tw.Close()).To(Succeed())

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, buffer.String())

			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))
	})

	it.After(func() {
		mockServer.Close()
	})

	context("given a dependency URL to get the license for", func() {
		it("gets the artifact and retrieves the license from it", func() {
			command := exec.Command(
				entrypoint,
				"--dependency-name", "dependency",
				"--url", fmt.Sprintf("%s/default-dependency-source-url.tgz", mockServer.URL),
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("Getting the dependency artifact from %s/default-dependency-source-url.tgz", mockServer.URL)))

			Expect(buffer).To(gbytes.Say("Decompressing the dependency artifact"))
			Expect(buffer).To(gbytes.Say("Scanning artifact for license file"))

			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::\[MIT MIT-0\]`))
			Expect(buffer).To(gbytes.Say("Licenses found!"))
		})
	})

	context("given a dotnet-runtime dependency URL to get the license for", func() {
		it("gets the artifact and retrieves the license from it", func() {
			command := exec.Command(
				entrypoint,
				"--dependency-name", "dotnet-runtime",
				"--url", fmt.Sprintf("%s/dotnet-source-url.tgz", mockServer.URL),
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("Getting the dependency artifact from %s/dotnet-source-url.tgz", mockServer.URL)))

			Expect(buffer).To(gbytes.Say("Decompressing the dependency artifact"))
			Expect(buffer).To(gbytes.Say("Scanning artifact for license file"))

			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::\[MIT MIT-0\]`))
			Expect(buffer).To(gbytes.Say("Licenses found!"))
		})
	})

	context("given a dotnet-aspnetcore dependency URL to get the license for", func() {
		it("gets the artifact and retrieves the license from it", func() {
			command := exec.Command(
				entrypoint,
				"--dependency-name", "dotnet-aspnetcore",
				"--url", fmt.Sprintf("%s/dotnet-source-url.tgz", mockServer.URL),
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("Getting the dependency artifact from %s/dotnet-source-url.tgz", mockServer.URL)))

			Expect(buffer).To(gbytes.Say("Decompressing the dependency artifact"))
			Expect(buffer).To(gbytes.Say("Scanning artifact for license file"))

			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::\[MIT MIT-0\]`))
			Expect(buffer).To(gbytes.Say("Licenses found!"))
		})
	})

	context("given a dotnet-sdk dependency URL to get the license for", func() {
		it("gets the artifact and retrieves the license from it", func() {
			command := exec.Command(
				entrypoint,
				"--dependency-name", "dotnet-sdk",
				"--url", fmt.Sprintf("%s/dotnet-source-url.tgz", mockServer.URL),
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("Getting the dependency artifact from %s/dotnet-source-url.tgz", mockServer.URL)))

			Expect(buffer).To(gbytes.Say("Decompressing the dependency artifact"))
			Expect(buffer).To(gbytes.Say("Scanning artifact for license file"))

			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::\[MIT MIT-0\]`))
			Expect(buffer).To(gbytes.Say("Licenses found!"))
		})
	})

	context("failure cases", func() {
		context("the --dependency-name flag is missing", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--url", fmt.Sprintf("%s/default-dependency-source-url.tgz", mockServer.URL),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`missing required flag --dependency-name`))
			})
		})

		context("the --url flag is missing", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "dependency",
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`missing required flag --url`))
			})
		})

		context("the request to the source URL fails", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "dependency",
					"--url", "non-existent/url",
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`failed to query url: Get "non-existent/url"`))
			})
		})

		context("the status code of the response is not OK", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "dependency",
					"--url", fmt.Sprintf("%s/bad-url", mockServer.URL),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(fmt.Sprintf(`failed to query url %s/bad-url with: status code 400`, mockServer.URL)))
			})
		})

		context("the artifact cannot be decompressed", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "dependency",
					"--url", fmt.Sprintf("%s/non-tar-file-artifact", mockServer.URL),
				)

				buffer := gbytes.NewBuffer()
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(gexec.Exit(1), func() string { return string(buffer.Contents()) })
				Expect(buffer).To(gbytes.Say(`failed to decompress source file`))
			})
		})

		context("the artifact does not contain a license", func() {
			it("returns an error and exits non-zero", func() {
				command := exec.Command(
					entrypoint,
					"--dependency-name", "dependency",
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
