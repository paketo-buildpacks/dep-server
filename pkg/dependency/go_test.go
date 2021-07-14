package dependency_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
)

func TestGo(t *testing.T) {
	spec.Run(t, "Go", testGo, spec.Report(report.Terminal{}))
}

func testGo(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		golang               dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}

		var err error
		golang, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient, fakeLicenseRetriever).NewDependency("go")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all go release versions with newest versions first", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
<!DOCTYPE html>
<html lang="en">
	<h2 id="go1.14">go1.14 (released 2020-02-25)</h2>
		<p>
		go1.14.1
		(released 2020-03-19)
		</p>
	<h2 id="go1.13">go1.13 (released 2019-09-03)</h2>
		<p>
		go1.13.8
		(released 2020-02-12)
		</p>
		<p>
		go1.13.9
		(released 2020-03-19)
		</p>
`), nil)

			fakeWebClient.GetReturnsOnCall(1, []byte(`
[
  {"version": "go1.14.1", "files": [{"sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "kind": "source"}]},
  {"version": "go1.13.9", "files": [{"sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "kind": "source"}]},
  {"version": "go1.14", "files": [{"sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", "kind": "source"}]},
  {"version": "go1.13.8", "files": [{"sha256": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", "kind": "source"}]},
  {"version": "go1.13", "files": [{"sha256": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", "kind": "source"}]}
]`), nil)

			versions, err := golang.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"1.14.1", "1.13.9", "1.14.0", "1.13.8", "1.13.0"}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://golang.org/doc/devel/release.html", urlArg)
		})

		it("ignores versions that can't be downloaded", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
<!DOCTYPE html>
<html lang="en">
	<h2 id="go1.13">go1.13 (released 2019-09-03)</h2>
		<p>
		go1.13.8
		(released 2020-02-12)
		</p>
		<p>
	<h2 id="go1.2">go1.2 (released 2013-12-01)</h2>
`), nil)

			fakeWebClient.GetReturnsOnCall(1, []byte(`
[
 {"version": "go1.13", "files": [{"sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "kind": "source"}]},
 {"version": "go1.13.8", "files": [{"sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "kind": "source"}]}
]`), nil)

			versions, err := golang.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"1.13.8", "1.13.0"}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://golang.org/doc/devel/release.html", urlArg)

			urlArg, _ = fakeWebClient.GetArgsForCall(1)
			assert.Equal("https://golang.org/dl/?mode=json&include=all", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		var expectedReleaseDate = time.Date(2020, 03, 19, 0, 0, 0, 0, time.UTC)

		it("returns the correct go version", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
[
 {"version": "go1.14.1", "files": [{"sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "kind": "source"}]},
 {"version": "go1.13.9", "files": [{"sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "kind": "source"}]},
 {"version": "go1.14", "files": [{"sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", "kind": "source"}]}
]`), nil)

			fakeWebClient.GetReturnsOnCall(1, []byte(`
<!DOCTYPE html>
<html lang="en">
	<h2 id="go1.14">go1.14 (released 2020-02-25)</h2>
		<p>
		go1.14.1
		(released 2020-03-19)
		</p>
	<h2 id="go1.13">go1.13 (released 2019-09-03)</h2>
		<p>
		go1.13.1
		(released 2019-09-25)
		</p>
		<p>
		go1.13.2
		(released 2019-10-17)
		</p>
		<p>
		go1.13.9
		(released 2020-03-19)
		</p>
`), nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

			actualDep, err := golang.GetDependencyVersion("go1.13.9")
			require.NoError(err)

			expectedDep := dependency.DepVersion{
				Version:         "1.13.9",
				URI:             "https://dl.google.com/go/go1.13.9.src.tar.gz",
				SHA256:          "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:golang:go:1.13.9:*:*:*:*:*:*:*",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://golang.org/dl/?mode=json&include=all", urlArg)

			urlArg, _ = fakeWebClient.GetArgsForCall(1)
			assert.Equal("https://golang.org/doc/devel/release.html", urlArg)
		})

		when("the SHA256 is empty", func() {
			it("calculates the SHA256", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
[
{"version": "go1.13.9", "files": [{"sha256": "", "kind": "source"}]}
]`), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte(`
<!DOCTYPE html>
<html lang="en">
	<h2 id="go1.13">go1.13 (released 2019-09-03)</h2>
		<p>
		go1.13.9
		(released 2020-03-19)
		</p>
`), nil)

				fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

				actualDep, err := golang.GetDependencyVersion("go1.13.9")
				require.NoError(err)

				expectedDep := dependency.DepVersion{
					Version:         "1.13.9",
					URI:             "https://dl.google.com/go/go1.13.9.src.tar.gz",
					SHA256:          "some-source-sha",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:golang:go:1.13.9:*:*:*:*:*:*:*",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://golang.org/dl/?mode=json&include=all", urlArg)

				urlArg, dependencyPathDownloadArg, _ := fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("https://dl.google.com/go/go1.13.9.src.tar.gz", urlArg)

				assert.Equal(dependencyPathDownloadArg, fakeChecksummer.GetSHA256ArgsForCall(0))
			})
		})

		when("a source file cannot be found", func() {
			it("returns a NoSourceCode error", func() {
				fakeWebClient.GetReturns([]byte(`
[
 {"version": "go1.14", "files": [{"sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "kind": "archive"}]}
]`), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte(`
<!DOCTYPE html>
<html lang="en">
	<h2 id="go1.14">go1.14 (released 2020-02-25)</h2>
		<p>
		go1.14.1
		(released 2020-03-19)
		</p>
`), nil)

				_, err := golang.GetDependencyVersion("go1.14")
				assert.Error(err)

				assert.True(errors.Is(err, depErrors.NoSourceCodeError{Version: "go1.14"}))
			})
		})
	})

	when("GetReleaseDate", func() {
		it.Before(func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
<!DOCTYPE html>
<html lang="en">
	<h2 id="go1.14">go1.14 (released 2020-02-25)</h2>
		<p>
		go1.14.1
		(released 2020/03/19)
		</p>
	<h2 id="go1.13">go1.13 (released 2019-09/03)</h2>
		<p>
		go1.13.1
		(released 2019-09-25)
		</p>
		<p>
		go1.13.2
		(released 2019-10-17)
		</p>
		<p>
		go1.13.9
		(released 2020-03-19)
		</p>
`), nil)
		})

		it("returns the correct release date", func() {
			releaseDate, err := golang.GetReleaseDate("go1.13.1")
			require.NoError(err)

			assert.Equal("2019-09-25T00:00:00Z", releaseDate.Format(time.RFC3339))
		})

		when("the release date cannot be found", func() {
			it("returns an error", func() {
				_, err := golang.GetReleaseDate("go9.99.999")
				assert.Error(err)

				assert.Equal("could not find release date", err.Error())
			})
		})
	})
}
