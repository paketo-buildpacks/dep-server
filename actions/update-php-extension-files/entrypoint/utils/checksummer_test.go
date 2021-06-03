package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestChecksummer(t *testing.T) {
	spec.Run(t, "checksummer", testChecksummer, spec.Report(report.Terminal{}))
}

func testChecksummer(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect      = NewWithT(t).Expect
		checksummer utils.ChecksummerPHP
	)

	it.Before(func() {
		checksummer = utils.Checksummer{}
	})

	context("GetMD5", func() {
		var (
			md5File string
		)
		it.Before(func() {
			workingDir, err := os.Getwd()
			md5File = filepath.Join(workingDir, "testdata", "example.txt")
			Expect(err).NotTo(HaveOccurred())
		})

		it("calculates the MD5 checksum of a given file", func() {
			expectedMD5 := "9058c04d83e6715d15574b1b51fadba8"

			depMD5, err := checksummer.GetMD5(md5File)
			Expect(err).NotTo(HaveOccurred())
			Expect(depMD5).To(Equal(expectedMD5))
		})

		context("failure cases", func() {
			context("when data file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Chmod(md5File, 0000)).To(Succeed())
				})
				it.After(func() {
					Expect(os.Chmod(md5File, 0664)).To(Succeed())
				})
				it("return an error", func() {
					_, err := checksummer.GetMD5(md5File)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

}
