package purl_test

import (
	"testing"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/purl"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestPURLGenerator(t *testing.T) {
	spec.Run(t, "PURLGenerator", testPURLGenerator, spec.Report(report.Terminal{}))
}

func testPURLGenerator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect        = NewWithT(t).Expect
		purlGenerator purl.PURLGenerator
	)

	context("Generate", func() {
		it.Before(func() {
			purlGenerator = purl.NewPURLGenerator()
		})

		it("returns a PURL", func() {
			purl := purlGenerator.Generate("dependencyName", "dependencyVersion", "dependencySourceSHA", "http://dependencySource")
			Expect(purl).To(Equal("pkg:generic/dependencyName@dependencyVersion?checksum=dependencySourceSHA&download_url=http://dependencySource"))
		})

	})
}
