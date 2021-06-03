package main_test

import (
	"fmt"
	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils"
	"testing"

	main "github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint"
	"github.com/paketo-buildpacks/dep-server/actions/update-php-extension-files/entrypoint/utils/utilsfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	spec.Run(t, "MainTest", mainTest, spec.Report(report.Terminal{}), spec.Parallel())
}

func mainTest(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect   = NewWithT(t).Expect
		phpUtils *utilsfakes.FakePHPExtUtils
	)

	it.Before(func() {
		phpUtils = &utilsfakes.FakePHPExtUtils{}
	})

	context("GenerateJSONPayload", func() {
		var (
			phpYMLFilesDir string
		)

		it.Before(func() {
			// create a temp dir
			// create example yml files on-the-fly
			phpUtils.GetPHPExtensionsYMLFilesReturns(map[string]utils.PHPExtMetadataFile{
				"php-7": {
					NativeModules: []utils.PHPExtension{
						{
							Name:    "1",
							Version: "1",
							MD5:     "1",
							Klass:   "1",
						},
						{
							Name:    "2",
							Version: "2",
							MD5:     "2",
							Klass:   "2",
						},
					},
					Extensions: []utils.PHPExtension{
						{
							Name:    "3",
							Version: "3",
							MD5:     "3",
							Klass:   "3",
						},
						{
							Name:    "4",
							Version: "4",
							MD5:     "4",
							Klass:   "4",
						},
					},
				},
				"php-8": {
					NativeModules: []utils.PHPExtension{
						{
							Name:    "5",
							Version: "5",
							MD5:     "5",
							Klass:   "5",
						},
						{
							Name:    "6",
							Version: "6",
							MD5:     "6",
							Klass:   "6",
						},
					},
					Extensions: []utils.PHPExtension{
						{
							Name:    "7",
							Version: "7",
							MD5:     "7",
							Klass:   "7",
						},
						{
							Name:    "8",
							Version: "8",
							MD5:     "8",
							Klass:   "8",
						},
					},
				},
			}, nil)

			phpUtils.GetUpdatedMetadataFileReturnsOnCall(0, utils.PHPExtMetadataFile{
				NativeModules: []utils.PHPExtension{},
				Extensions: []utils.PHPExtension{
					{
						Name:    "3",
						Version: "3-UPDATED",
						MD5:     "3-MD5-UPDATED",
						Klass:   "3",
					},
				},
			}, nil)

			phpUtils.GetUpdatedMetadataFileReturnsOnCall(1, utils.PHPExtMetadataFile{
				NativeModules: []utils.PHPExtension{
					{
						Name:    "6",
						Version: "6-UPDATED",
						MD5:     "6-MD5-UPDATED",
						Klass:   "6",
					},
				},
				Extensions: []utils.PHPExtension{
					{
						Name:    "8",
						Version: "8-UPDATED",
						MD5:     "8-MD5-UPDATED",
						Klass:   "8",
					},
				},
			}, nil)

		})

		it("", func() {
			expectedJson := `{"data":{"php-7":{"NativeModules":[],"Extensions":[{"Name":"3","Version":"3-UPDATED","MD5":"3-MD5-UPDATED","Klass":"3"}]},"php-8":{"NativeModules":[{"Name":"6","Version":"6-UPDATED","MD5":"6-MD5-UPDATED","Klass":"6"}],"Extensions":[{"Name":"8","Version":"8-UPDATED","MD5":"8-MD5-UPDATED","Klass":"8"}]}}}`

			jsonObj, err := main.GenerateJSONPayload(phpUtils, phpYMLFilesDir)
			fmt.Println(jsonObj)
			fmt.Println(expectedJson)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonObj).To(Equal(expectedJson))
		})

	})
}
