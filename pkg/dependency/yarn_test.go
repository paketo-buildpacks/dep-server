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
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
)

func TestYarn(t *testing.T) {
	spec.Run(t, "Yarn", testYarn, spec.Report(report.Terminal{}))
}

func testYarn(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		yarn                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		yarn, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("yarn")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all yarn final release versions with newest versions first", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "v3.0.0",
					PublishedDate: time.Date(2020, 6, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v1.0.1",
					PublishedDate: time.Date(2020, 6, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v2.0.0",
					PublishedDate: time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v2.0.0-exp.1",
					PublishedDate: time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
				},
			}, nil)

			versions, err := yarn.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"3.0.0", "1.0.1", "2.0.0", "1.0.0"}, versions)

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
					PublishedDate: time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
				},
			}, nil)
			assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
			fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)

			fakeGithubClient.GetReleaseAssetReturns([]byte("some-signature"), nil)
			fakeGithubClient.DownloadReleaseAssetReturns("some-asset-url", nil)
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/yarn@1.0.0?checksum=some-source-sha&download_url=some-source-url")

			actualDep, err := yarn.GetDependencyVersion("1.0.0")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:         "1.0.0",
				URI:             "some-source-url",
				SHA256:          "some-source-sha",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:yarnpkg:yarn:1.0.0:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/yarn@1.0.0?checksum=some-source-sha&download_url=some-source-url",
				Licenses:        []string{"MIT", "MIT-2"},
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
						PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
				assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
				fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)
				fakeGithubClient.DownloadReleaseAssetReturns("", internal_errors.AssetNotFound{AssetName: "yarn-v1.0.0.tar.gz"})

				_, err := yarn.GetDependencyVersion("1.0.0")
				assert.Error(err)

				assert.True(errors.Is(err, depErrors.NoSourceCodeError{Version: "1.0.0"}))
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct yarn release date", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "v2.0.0",
					PublishedDate: time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
				},
			}, nil)

			releaseDate, err := yarn.GetReleaseDate("v1.0.0")
			require.NoError(err)

			assert.Equal("2020-06-27T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
