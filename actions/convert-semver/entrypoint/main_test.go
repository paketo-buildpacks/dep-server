package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/sclevine/spec"

	"github.com/sclevine/spec/report"
	assertpkg "github.com/stretchr/testify/assert"
	requirepkg "github.com/stretchr/testify/require"
)

func TestEntrypoint(t *testing.T) {
	spec.Run(t, "Entrypoint", testEntrypoint, spec.Report(report.Terminal{}))
}

func testEntrypoint(t *testing.T, when spec.G, it spec.S) {
	var (
		assert  = assertpkg.New(t)
		require = requirepkg.New(t)
		cliPath string
	)

	it.Before(func() {
		tempFile, err := ioutil.TempFile("", "entrypoint")
		require.NoError(err)
		cliPath = tempFile.Name()
		require.NoError(tempFile.Close())

		goBuild := exec.Command("go", "build", "-o", cliPath, ".")
		output, err := goBuild.CombinedOutput()
		require.NoError(err, "failed to build CLI: %s", string(output))
	})

	it.After(func() {
		_ = os.Remove(cliPath)
	})

	it("returns semantic version of the specified Go dependency version", func() {
		version := "go1.15"

		actualVersion, err := exec.Command(cliPath, "--version", version).CombinedOutput()
		require.NoError(err, string(actualVersion))
		assert.Equal("1.15.0", string(actualVersion))
	})

	when("given a major.minor version", func() {
		it("returns the semantic version", func() {
			version := "1.15"

			actualVersion, err := exec.Command(cliPath, "--version", version).CombinedOutput()
			require.NoError(err, string(actualVersion))
			assert.Equal("1.15.0", string(actualVersion))
		})
	})

	when("given a semantic version", func() {
		it("returns the same version", func() {
			version := "1.15.1"

			actualVersion, err := exec.Command(cliPath, "--version", version).CombinedOutput()
			require.NoError(err, string(actualVersion))
			assert.Equal("1.15.1", string(actualVersion))
		})
	})

	when("failure cases", func() {
		when("the version cannot be converted into semver", func() {
			it("exits status 1", func() {
				version := "abc%"
				_, err := exec.Command(cliPath, "--version", version).CombinedOutput()
				assert.Error(err)
			})
		})
	})

}
