package dependency_test

import (
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTini(t *testing.T) {
	spec.Run(t, "Tini", testTini, spec.Report(report.Terminal{}))
}

func testTini(t *testing.T, when spec.G, it spec.S) {
	var (
		assert           = assert.New(t)
		require          = require.New(t)
		fakeChecksummer  *dependencyfakes.FakeChecksummer
		fakeGithubClient *dependencyfakes.FakeGithubClient
		tini             dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}

		var err error
		tini, err = dependency.NewCustomDependencyFactory(fakeChecksummer, nil, fakeGithubClient, nil).NewDependency("tini")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all tini release versions with newest versions first", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "v3.0.0",
					PublishedDate: "2020-06-30T00:00:00Z",
				},
				{
					TagName:       "v1.0.1",
					PublishedDate: "2020-06-29T00:00:00Z",
				},
				{
					TagName:       "v2.0.0",
					PublishedDate: "2020-06-28T00:00:00Z",
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: "2020-06-27T00:00:00Z",
				},
			}, nil)

			versions, err := tini.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"v3.0.0", "v1.0.1", "v2.0.0", "v1.0.0"}, versions)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("krallin", orgArg)
			assert.Equal("tini", repoArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct tini version", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "v2.0.0",
					PublishedDate: "2020-06-28T00:00:00Z",
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: "2020-06-27T00:00:00Z",
				},
			}, nil)
			fakeGithubClient.DownloadSourceTarballReturns("some-tarball-url", nil)
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)

			actualDep, err := tini.GetDependencyVersion("v1.0.0")
			require.NoError(err)

			expectedDep := dependency.DepVersion{
				Version:         "v1.0.0",
				URI:             "some-tarball-url",
				SHA:             "some-source-sha",
				ReleaseDate:     "2020-06-27T00:00:00Z",
				DeprecationDate: "",
			}

			assert.Equal(expectedDep, actualDep)

			orgArg, repoArg, versionArg, _ := fakeGithubClient.DownloadSourceTarballArgsForCall(0)
			assert.Equal("krallin", orgArg)
			assert.Equal("tini", repoArg)
			assert.Equal("v1.0.0", versionArg)
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct tini release date", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "v2.0.0",
					PublishedDate: "2020-06-28T00:00:00Z",
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: "2020-06-27T00:00:00Z",
				},
			}, nil)

			releaseDate, err := tini.GetReleaseDate("v1.0.0")
			require.NoError(err)

			assert.Equal("2020-06-27T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
