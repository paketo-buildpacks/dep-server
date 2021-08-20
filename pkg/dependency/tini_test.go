package dependency_test

import (
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
)

func TestTini(t *testing.T) {
	spec.Run(t, "Tini", testTini, spec.Report(report.Terminal{}))
}

func testTini(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		tini                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		tini, err = dependency.NewCustomDependencyFactory(fakeChecksummer, nil, fakeGithubClient, nil, fakeLicenseRetriever, fakePURLGenerator).NewDependency("tini")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all tini release versions with newest versions first", func() {
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
					TagName:       "v1.0.0",
					PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
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
					PublishedDate: time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
				},
			}, nil)
			fakeGithubClient.DownloadSourceTarballReturns("some-tarball-url", nil)

			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/tini@v1.0.0?checksum=some-source-sha&download_url=some-tarball-url")

			actualDep, err := tini.GetDependencyVersion("v1.0.0")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:         "v1.0.0",
				URI:             "some-tarball-url",
				SHA256:          "some-source-sha",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:tini_project:tini:1.0.0:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/tini@v1.0.0?checksum=some-source-sha&download_url=some-tarball-url",
				Licenses:        []string{"MIT", "MIT-2"},
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
					PublishedDate: time.Date(2020, 6, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC),
				},
			}, nil)

			releaseDate, err := tini.GetReleaseDate("v1.0.0")
			require.NoError(err)

			assert.Equal("2020-06-27T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
