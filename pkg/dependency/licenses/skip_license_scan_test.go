package licenses_test

import (
	"testing"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/licenses"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSkipLicenseScanCase(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect           = NewWithT(t).Expect
		licenseRetriever licenses.LicenseRetriever
	)

	context("given the composer dependency", func() {
		it.Before(func() {
			licenseRetriever = licenses.NewLicenseRetriever()
		})
		it("the function exits", func() {
			licenses, err := licenseRetriever.LookupLicenses("composer", "composer-dep-server-url")
			Expect(err).NotTo(HaveOccurred())
			Expect(licenses).To(Equal([]string{}))
		})
	})
}
