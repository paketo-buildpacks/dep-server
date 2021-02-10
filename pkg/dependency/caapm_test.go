package dependency_test

import (
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCAAPM(t *testing.T) {
	spec.Run(t, "CAAPM", testCAAPM, spec.Report(report.Terminal{}))
}

func testCAAPM(t *testing.T, when spec.G, it spec.S) {
	var (
		assert          = assert.New(t)
		require         = require.New(t)
		fakeChecksummer *dependencyfakes.FakeChecksummer
		fakeFileSystem  *dependencyfakes.FakeFileSystem
		fakeWebClient   *dependencyfakes.FakeWebClient
		caapm           dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}

		var err error
		caapm, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient).NewDependency("CAAPM")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all caapm release versions with highest versions first", func() {
			fakeWebClient.GetReturns([]byte(`
<html>
<head>
</head>
<body>
<pre><a href="AAA-WRONGAgent-1.0.1_linux.tar.gz">AAA-WRONGAgent-1.0.0_linux.tar.gz</a></pre>
<pre><a href="CA-APM-PHPAgent-10.6.0_linux.tar.gz">CA-APM-PHPAgent-10.6.0_linux.tar.gz</a></pre>
<pre><a href="CA-APM-PHPAgent-10.7.0_linux.tar.gz">CA-APM-PHPAgent-10.7.0_linux.tar.gz</a></pre>
<pre><a href="CA-APM-PHPAgent-20.1.0_linux.tar.gz">CA-APM-PHPAgent-20.1.0_linux.tar.gz</a></pre>
<pre><a href="CA-APM-PHPAgent-9.9.9_linux.tar.gz">CA-APM-PHPAgent-9.9.9_linux.tar.gz</a></pre>
<pre><a href="CA-APM-WRONGAgent-99.99.99_linux.tar.gz">CA-APM-WRONGAgent-99.99.99_linux.tar.gz</a></pre>
</body>
</html>
`), nil)

			versions, err := caapm.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{
				"20.1.0",
				"10.7.0",
				"10.6.0",
				"9.9.9",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://ca.bintray.com/apm-agents/", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct caapm version", func() {
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)

			actualDepVersion, err := caapm.GetDependencyVersion("20.1.0")
			require.NoError(err)

			expectedDepVersion := dependency.DepVersion{
				Version:         "20.1.0",
				URI:             "https://ca.bintray.com/apm-agents/CA-APM-PHPAgent-20.1.0_linux.tar.gz",
				SHA:             "some-source-sha",
				ReleaseDate:     "",
				DeprecationDate: "",
			}

			assert.Equal(expectedDepVersion, actualDepVersion)

			dependencyURLArg, _, _ := fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("https://ca.bintray.com/apm-agents/CA-APM-PHPAgent-20.1.0_linux.tar.gz", dependencyURLArg)
		})
	})

	when("GetReleaseDate", func() {
		it("returns an error", func() {
			_, err := caapm.GetReleaseDate("20.1.0")
			require.Error(err)
			assert.Equal("cannot determine release dates for CAAPM", err.Error())
		})
	})
}
