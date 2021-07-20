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

func TestICU(t *testing.T) {
	spec.Run(t, "ICU", testICU, spec.Report(report.Terminal{}))
}

func testICU(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		icu                  dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		icu, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("icu")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all ICU release versions with newest versions first", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{TagName: "release-59-1", CreatedDate: time.Date(2017, 04, 13, 13, 46, 59, 0, time.UTC)},
				{TagName: "release-60-1", CreatedDate: time.Date(2017, 10, 31, 15, 14, 15, 0, time.UTC)},
				{TagName: "release-60-2", CreatedDate: time.Date(2017, 12, 13, 20, 01, 38, 0, time.UTC)},
				{TagName: "release-67-1", CreatedDate: time.Date(2020, 04, 22, 17, 49, 10, 0, time.UTC)},
				{TagName: "release-66-1", CreatedDate: time.Date(2020, 03, 11, 17, 21, 07, 0, time.UTC)},
				{TagName: "release-66-1-99", CreatedDate: time.Date(2020, 03, 12, 17, 21, 07, 0, time.UTC)},
				{TagName: "release-65-1", CreatedDate: time.Date(2019, 10, 02, 21, 30, 54, 0, time.UTC)},
				{TagName: "release-60-3", CreatedDate: time.Date(2019, 04, 11, 18, 44, 36, 0, time.UTC)},
				{TagName: "release-59-2", CreatedDate: time.Date(2019, 04, 11, 18, 43, 42, 0, time.UTC)},
				{TagName: "release-4-8-2", CreatedDate: time.Date(2019, 04, 11, 18, 17, 52, 0, time.UTC)},
				{TagName: "release-99-99", CreatedDate: time.Date(2019, 04, 11, 18, 17, 52, 0, time.UTC)},
			}, nil)

			versions, err := icu.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{
				"67.1",
				"66.1.99",
				"66.1",
				"65.1",
				"60.3",
				"59.2",
				"99.99",
				"4.8.2",
				"60.2",
				"60.1",
			}, versions)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("unicode-org", orgArg)
			assert.Equal("icu", repoArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct ICU version", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{TagName: "release-67-1", CreatedDate: time.Date(2020, 04, 22, 17, 49, 10, 0, time.UTC)},
				{TagName: "release-66-1", CreatedDate: time.Date(2020, 03, 11, 17, 21, 07, 0, time.UTC)},
				{TagName: "release-65-1", CreatedDate: time.Date(2020, 10, 02, 21, 30, 54, 0, time.UTC)},
			}, nil)
			assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
			fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)
			fakeGithubClient.GetReleaseAssetReturns([]byte("some-signature"), nil)
			fakeGithubClient.DownloadReleaseAssetReturns("some-asset-url", nil)
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
			fakeChecksummer.SplitPGPKeysReturns([]string{"some-gpg-key"})
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/icu@66.1?checksum=some-source-sha&download_url=some-source-url")

			actualDep, err := icu.GetDependencyVersion("66.1")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 03, 11, 17, 21, 07, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:         "66.1",
				URI:             "some-source-url",
				SHA256:          "some-source-sha",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             `cpe:2.3:a:icu-project:international_components_for_unicode:66.1:*:*:*:*:c\/c\+\+:*:*`,
				PURL:            "pkg:generic/icu@66.1?checksum=some-source-sha&download_url=some-source-url",
				Licenses:        []string{"MIT", "MIT-2"},
			}

			assert.Equal(expectedDep, actualDep)

			url, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://raw.githubusercontent.com/unicode-org/icu/master/KEYS", url)

			orgArg, repoArg, versionArg, filenameArg, _ := fakeGithubClient.DownloadReleaseAssetArgsForCall(0)
			assert.Equal("unicode-org", orgArg)
			assert.Equal("icu", repoArg)
			assert.Equal("release-66-1", versionArg)
			assert.Equal("icu4c-66_1-src.tgz", filenameArg)

			orgArg, repoArg, versionArg, filenameArg = fakeGithubClient.GetReleaseAssetArgsForCall(0)
			assert.Equal("unicode-org", orgArg)
			assert.Equal("icu", repoArg)
			assert.Equal("release-66-1", versionArg)
			assert.Equal("icu4c-66_1-src.tgz.asc", filenameArg)

			assert.Equal("some-gpg-key", fakeChecksummer.SplitPGPKeysArgsForCall(0))

			releaseAssetSignatureArg, _, gpgKeysArg := fakeChecksummer.VerifyASCArgsForCall(0)
			assert.Equal("some-signature", releaseAssetSignatureArg)
			assert.Equal([]string{"some-gpg-key"}, gpgKeysArg)
		})

		when("getting a version prior to 49", func() {
			it("returns the correct ICU version", func() {
				fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
					{TagName: "release-4-8-3", CreatedDate: time.Date(2019, 04, 12, 18, 17, 52, 0, time.UTC)},
					{TagName: "release-4-8-2", CreatedDate: time.Date(2019, 04, 11, 18, 17, 52, 0, time.UTC)},
					{TagName: "release-4-8-1", CreatedDate: time.Date(2019, 04, 10, 18, 17, 52, 0, time.UTC)},
				}, nil)
				assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
				fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)
				fakeGithubClient.GetReleaseAssetReturns([]byte("some-signature"), nil)
				fakeGithubClient.DownloadReleaseAssetReturns("some-asset-url", nil)
				fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/icu@4.8.2?checksum=some-source-sha&download_url=some-source-url")

				actualDep, err := icu.GetDependencyVersion("4.8.2")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedReleaseDate := time.Date(2019, 04, 11, 18, 17, 52, 0, time.UTC)
				expectedDep := dependency.DepVersion{
					Version:         "4.8.2",
					URI:             "some-source-url",
					SHA256:          "some-source-sha",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             `cpe:2.3:a:icu-project:international_components_for_unicode:4.8.2:*:*:*:*:c\/c\+\+:*:*`,
					PURL:            "pkg:generic/icu@4.8.2?checksum=some-source-sha&download_url=some-source-url",
					Licenses:        []string{"MIT", "MIT-2"},
				}

				assert.Equal(expectedDep, actualDep)

				url, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://raw.githubusercontent.com/unicode-org/icu/master/KEYS", url)

				orgArg, repoArg, versionArg, filenameArg, _ := fakeGithubClient.DownloadReleaseAssetArgsForCall(0)
				assert.Equal("unicode-org", orgArg)
				assert.Equal("icu", repoArg)
				assert.Equal("release-4-8-2", versionArg)
				assert.Equal("icu4c-4_8_2-src.tgz", filenameArg)

				orgArg, repoArg, versionArg, filenameArg = fakeGithubClient.GetReleaseAssetArgsForCall(0)
				assert.Equal("unicode-org", orgArg)
				assert.Equal("icu", repoArg)
				assert.Equal("release-4-8-2", versionArg)
				assert.Equal("icu4c-4_8_2-src.tgz.asc", filenameArg)
			})
		})

		when("the asset cannot be found", func() {
			it("returns a NoSourceCode error", func() {
				fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
					{TagName: "release-66-1", CreatedDate: time.Date(2020, 03, 11, 17, 21, 07, 0, time.UTC)},
				}, nil)
				assetUrlContent := `{"browser_download_url":"some-source-url", "key":"some_value"}`
				fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(assetUrlContent), nil)
				fakeGithubClient.DownloadReleaseAssetReturns("", internal_errors.AssetNotFound{AssetName: "icu4c-66_1-src.tgz"})

				_, err := icu.GetDependencyVersion("66.1")
				assert.Error(err)

				assert.True(errors.Is(err, depErrors.NoSourceCodeError{Version: "66.1"}))
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct ICU release date", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{TagName: "release-67-1", CreatedDate: time.Date(2020, 04, 22, 17, 49, 10, 0, time.UTC)},
				{TagName: "release-66-1", CreatedDate: time.Date(2020, 03, 11, 17, 21, 07, 0, time.UTC)},
				{TagName: "release-65-1", CreatedDate: time.Date(2019, 10, 02, 21, 30, 54, 0, time.UTC)},
			}, nil)

			releaseDate, err := icu.GetReleaseDate("66.1")
			require.NoError(err)

			assert.Equal("2020-03-11T17:21:07Z", releaseDate.Format(time.RFC3339))
		})
	})
}
