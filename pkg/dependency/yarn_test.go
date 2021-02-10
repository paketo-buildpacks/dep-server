package dependency_test

import (
	"errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestYarn(t *testing.T) {
	spec.Run(t, "Yarn", testYarn, spec.Report(report.Terminal{}))
}

func testYarn(t *testing.T, when spec.G, it spec.S) {
	var (
		assert           = assert.New(t)
		require          = require.New(t)
		fakeChecksummer  *dependencyfakes.FakeChecksummer
		fakeFileSystem   *dependencyfakes.FakeFileSystem
		fakeGithubClient *dependencyfakes.FakeGithubClient
		fakeWebClient    *dependencyfakes.FakeWebClient
		yarn             dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}

		var err error
		yarn, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient).NewDependency("yarn")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all yarn release versions with newest versions first", func() {
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

			versions, err := yarn.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"v3.0.0", "v1.0.1", "v2.0.0", "v1.0.0"}, versions)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("yarnpkg", orgArg)
			assert.Equal("yarn", repoArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct yarn version", func() {
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
			assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
			fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)
			fakeGithubClient.GetReleaseAssetReturns([]byte("some-signature"), nil)
			fakeGithubClient.DownloadReleaseAssetReturns("some-asset-url", nil)
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)

			actualDep, err := yarn.GetDependencyVersion("v1.0.0")
			require.NoError(err)

			expectedDep := dependency.DepVersion{
				Version:         "v1.0.0",
				URI:             "some-source-url",
				SHA:             "some-source-sha",
				ReleaseDate:     "2020-06-27T00:00:00Z",
				DeprecationDate: "",
			}

			assert.Equal(expectedDep, actualDep)

			url, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://dl.yarnpkg.com/debian/pubkey.gpg", url)

			orgArg, repoArg, versionArg, filenameArg := fakeGithubClient.GetReleaseAssetArgsForCall(0)
			assert.Equal("yarnpkg", orgArg)
			assert.Equal("yarn", repoArg)
			assert.Equal("v1.0.0", versionArg)
			assert.Equal("yarn-v1.0.0.tar.gz.asc", filenameArg)

			orgArg, repoArg, versionArg, filenameArg, _ = fakeGithubClient.DownloadReleaseAssetArgsForCall(0)
			assert.Equal("yarnpkg", orgArg)
			assert.Equal("yarn", repoArg)
			assert.Equal("v1.0.0", versionArg)
			assert.Equal("yarn-v1.0.0.tar.gz", filenameArg)

			releaseAssetSignatureArg, _, yarnGPGKeyArg := fakeChecksummer.VerifyASCArgsForCall(0)
			assert.Equal("some-signature", releaseAssetSignatureArg)
			assert.Equal([]string{"some-gpg-key"}, yarnGPGKeyArg)
		})

		when("the asset cannot be found", func() {
			it("returns a NoSourceCode error", func() {
				fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
					{
						TagName:       "v1.0.0",
						PublishedDate: "2020-06-27T00:00:00Z",
					},
				}, nil)
				assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
				fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)
				fakeGithubClient.DownloadReleaseAssetReturns("", internal_errors.AssetNotFound{AssetName: "yarn-v1.0.0.tar.gz"})

				_, err := yarn.GetDependencyVersion("v1.0.0")
				assert.Error(err)

				assert.True(errors.Is(err, depErrors.NoSourceCodeError{Version: "v1.0.0"}))
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct yarn release date", func() {
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

			releaseDate, err := yarn.GetReleaseDate("v1.0.0")
			require.NoError(err)

			assert.Equal("2020-06-27T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
