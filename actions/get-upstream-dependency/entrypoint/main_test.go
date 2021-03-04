package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"

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

	it("returns formatted metadata of the version of the specified dependency", func() {
		version := "go1.15.6"

		output, err := exec.Command(cliPath, "--name", "go", "--version", version).CombinedOutput()
		require.NoError(err, string(output))

		var actualDepVersion dependency.DepVersion
		err = json.Unmarshal(output, &actualDepVersion)
		require.NoError(err)

		dep, err := dependency.NewDependencyFactory("").NewDependency("go")
		require.NoError(err)

		expectedDepVersion, err := dep.GetDependencyVersion(version)
		require.NoError(err)

		assert.Equal(expectedDepVersion, actualDepVersion)
	})
}
