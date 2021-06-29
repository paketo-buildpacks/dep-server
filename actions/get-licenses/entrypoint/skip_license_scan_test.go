package main_test

import (
	"os/exec"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSkipLicenseScanCase(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
	)

	context("given the CAAPM dependency", func() {
		it("the function exits", func() {
			command := exec.Command(
				entrypoint,
				"--dependency-name", "CAAPM",
				"--url", "caapm-dep-server-url",
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say("Skipping license retrieval for CAAPM"))
			Expect(buffer).To(gbytes.Say("License is not automatically retrievable and may need to be looked up manually"))
			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::nil`))
		})
	})

	context("given the composer dependency", func() {
		it("the function exits", func() {
			command := exec.Command(
				entrypoint,
				"--dependency-name", "composer",
				"--url", "composer-dep-server-url",
			)

			buffer := gbytes.NewBuffer()
			session, err := gexec.Start(command, buffer, buffer)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0), func() string { return string(buffer.Contents()) })

			Expect(buffer).To(gbytes.Say("Skipping license retrieval for composer"))
			Expect(buffer).To(gbytes.Say("License is not automatically retrievable and may need to be looked up manually"))
			Eventually(buffer).Should(gbytes.Say(`::set-output name=licenses::nil`))
		})
	})

}
