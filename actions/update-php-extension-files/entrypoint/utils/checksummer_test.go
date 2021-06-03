package utils_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestChecksummer(t *testing.T) {
	spec.Run(t, "checksummer", testChecksummer, spec.Report(report.Terminal{}))
}

func testChecksummer(t *testing.T, when spec.G, it spec.S) {

}
