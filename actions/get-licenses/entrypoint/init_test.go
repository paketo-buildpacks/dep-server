package main_test

import (
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	entrypoint string
)

func TestGetLicenses(t *testing.T) {
	Expect := NewWithT(t).Expect
	var err error
	entrypoint, err = gexec.Build("github.com/paketo-buildpacks/dep-server/actions/get-licenses/entrypoint")
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("getLicenses", spec.Report(report.Terminal{}))
	suite("DefaultLicenseCase", testDefaultLicenseCase)
	suite("BundlerLicenseCase", testBundlerLicenseCase)
	suite("SkipLicenseScanCase", testSkipLicenseScanCase)
	suite.Run(t)
}
