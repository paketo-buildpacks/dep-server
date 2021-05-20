package dependency_test

import (
	"testing"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithub(t *testing.T) {
	spec.Run(t, "Github", testGithub, spec.Report(report.Terminal{}))
}

func testGithub(t *testing.T, when spec.G, it spec.S) {
	var (
		assert           = assert.New(t)
		require          = require.New(t)
		fakeChecksummer  *dependencyfakes.FakeChecksummer
		fakeFileSystem   *dependencyfakes.FakeFileSystem
		fakeGithubClient *dependencyfakes.FakeGithubClient
		fakeWebClient    *dependencyfakes.FakeWebClient
		github           dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}

		var err error
		github, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient).NewDependency("alanxz/rabbitmq-c")
		require.NoError(err)
	})
	when("GetAllVersionRefs", func() {
		it("returns all final release versions with newest versions first", func() {
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

			versions, err := github.GetAllVersionRefs()
			require.NoError(err)
			assert.Equal([]string{"3.0.0", "1.0.1", "2.0.0", "1.0.0"}, versions)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("alanxz", orgArg)
			assert.Equal("rabbitmq-c", repoArg)
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

			actualDep, err := github.GetDependencyVersion("1.0.1")
			require.NoError(err)

			expectedReleaseDate := time.Date(2020, 6, 29, 0, 0, 0, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:         "1.0.1",
				URI:             "https://github.com/alanxz/rabbitmq-c/archive/v1.0.1.tar.gz",
				SHA256:          "",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
			}
			assert.Equal(expectedDep, actualDep)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("alanxz", orgArg)
			assert.Equal("rabbitmq-c", repoArg)
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

			releaseDate, err := github.GetReleaseDate("1.0.1")
			require.NoError(err)

			orgArg, repoArg := fakeGithubClient.GetReleaseTagsArgsForCall(0)
			assert.Equal("alanxz", orgArg)
			assert.Equal("rabbitmq-c", repoArg)

			assert.Equal("2020-06-29T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
