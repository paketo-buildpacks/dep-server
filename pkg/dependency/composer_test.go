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
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
)

func TestComposer(t *testing.T) {
	spec.Run(t, "composer", testComposer, spec.Report(report.Terminal{}))
}

func testComposer(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		composer             dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		composer, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("composer")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all composer final release versions with newest versions first", func() {
			time.Date(2020, 06, 30, 0, 0, 0, 0, time.UTC)
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "3.0.0",
					PublishedDate: time.Date(2020, 06, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "3.0.0-RC",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "3.0.0-beta1",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "3.0.0-alpha2",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "3.0.0-alpha1",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "1.0.1",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "2.0.0",
					PublishedDate: time.Date(2020, 06, 28, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "1.0.0",
					PublishedDate: time.Date(2020, 06, 27, 0, 0, 0, 0, time.UTC),
				},
			}, nil)

			versions, err := composer.GetAllVersionRefs()
			require.NoError(err)
			assert.Equal([]string{"3.0.0", "1.0.1", "2.0.0", "1.0.0"}, versions)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("composer", orgArg)
			assert.Equal("composer", repoArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct composer version", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "3.0.0",
					PublishedDate: time.Date(2020, 06, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "1.0.1",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "2.0.0",
					PublishedDate: time.Date(2020, 06, 28, 0, 0, 0, 0, time.UTC),
				},
			}, nil)
			fakeWebClient.GetReturnsOnCall(0,
				[]byte(`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  composer.phar`), nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/composer@1.0.1?checksum=aaaaaaaa&download_url=https://getcomposer.org")

			actualDep, err := composer.GetDependencyVersion("1.0.1")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 6, 29, 0, 0, 0, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:         "1.0.1",
				URI:             "https://getcomposer.org/download/1.0.1/composer.phar",
				SHA256:          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				PURL:            "pkg:generic/composer@1.0.1?checksum=aaaaaaaa&download_url=https://getcomposer.org",
				Licenses:        []string{},
			}
			assert.Equal(expectedDep, actualDep)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("composer", orgArg)
			assert.Equal("composer", repoArg)
		})

		when("the SHA256 cannot be found", func() {
			it("returns an error", func() {
				fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
					{
						TagName:       "3.0.0",
						PublishedDate: time.Date(2020, 06, 30, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
				fakeWebClient.GetReturnsOnCall(0, nil, nil)

				_, err := composer.GetDependencyVersion("3.0.0")
				assert.Error(err)

				assert.Contains(err.Error(), "could not get SHA256 from file")
			})
		})

		when("the licenses cannot be found", func() {
			it("returns an error", func() {
				fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
					{
						TagName:       "3.0.0",
						PublishedDate: time.Date(2020, 06, 30, 0, 0, 0, 0, time.UTC),
					},
					{
						TagName:       "1.0.1",
						PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
					},
					{
						TagName:       "2.0.0",
						PublishedDate: time.Date(2020, 06, 28, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
				fakeWebClient.GetReturnsOnCall(0,
					[]byte(`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  composer.phar`), nil)

				fakeLicenseRetriever.LookupLicensesReturns([]string{}, errors.New("failed licenses scan"))
				_, err := composer.GetDependencyVersion("3.0.0")
				assert.Error(err)

				assert.Contains(err.Error(), "could not find license metadata")
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct composer release date", func() {
			fakeGithubClient.GetReleaseTagsReturns([]internal.GithubRelease{
				{
					TagName:       "3.0.0",
					PublishedDate: time.Date(2020, 06, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "1.0.1",
					PublishedDate: time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC),
				},
				{
					TagName:       "2.0.0",
					PublishedDate: time.Date(2020, 06, 28, 0, 0, 0, 0, time.UTC),
				},
			}, nil)

			releaseDate, err := composer.GetReleaseDate("1.0.1")
			require.NoError(err)

			assert.Equal("2020-06-29T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
