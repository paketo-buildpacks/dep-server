package licenses_test

import (
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestGetLicenses(t *testing.T) {
	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("getLicenses", spec.Report(report.Terminal{}))
	suite("DefaultLicenseCase", testDefaultLicenseCase)
	suite("BundlerLicenseCase", testBundlerLicenseCase)
	suite("SkipLicenseScanCase", testSkipLicenseScanCase)
	suite.Run(t)
}
