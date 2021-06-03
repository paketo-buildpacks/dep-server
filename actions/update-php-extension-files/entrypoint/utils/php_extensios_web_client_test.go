package utils_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)
)

func TestPHPExtensionsWebClient(t *testing.T) {
	spec.Run(t, "PHPExtensionsWebClient", testPHPExtensionsWebClient, spec.Report(report.Terminal{}), spec.Parallel())
}

func testPHPExtensionsWebClient(t *testing.T, context spec.G, it spec.S) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		client  *httptest.Server
	)

	context("When using the testPHPExtensionsWebClient", func() {

		context("DownloadExtensionsSource", func() {

		})
	})

}
