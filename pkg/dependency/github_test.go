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
			fakeGithubClient.GetTagsReturns([]string{
				"1.49.0",
				"1.48.0",
				"1.47.0",
				"1.46.0-alpha1",
				"0.2",
				"release-0.1",
				"rabbitmq-c-0.3",
			}, nil)

			versions, err := github.GetAllVersionRefs()
			require.NoError(err)
			assert.Equal([]string{"1.49.0", "1.48.0", "1.47.0", "0.2"}, versions)

			orgArg, repoArg := fakeGithubClient.GetTagsArgsForCall(0)
			assert.Equal("alanxz", orgArg)
			assert.Equal("rabbitmq-c", repoArg)
		})
	})
	when("GetDependencyVersion", func() {
		it("returns the correct composer version", func() {
			date := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
			tagCommit := internal.GithubTagCommit{
				Tag:  "0.11.0",
				SHA:  "some-sha",
				Date: date,
			}
			fakeGithubClient.GetTagCommitReturns(tagCommit, nil)
			// fakeChecksummer.GetSHA256Returns("some-source-sha", nil)

			actualDepVersion, err := github.GetDependencyVersion("0.11.0")
			require.NoError(err)

			expectedDepVersion := dependency.DepVersion{
				Version:         "0.11.0",
				URI:             "https://github.com/alanxz/rabbitmq-c/archive/v0.11.0.tar.gz",
				SHA256:          "",
				ReleaseDate:     &tagCommit.Date,
				DeprecationDate: nil,
				CPE:             "",
			}
			assert.Equal(expectedDepVersion, actualDepVersion)
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct composer release date", func() {
			date := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
			tagCommit := internal.GithubTagCommit{
				Tag:  "0.11.0",
				SHA:  "some-sha",
				Date: date,
			}
			fakeGithubClient.GetTagCommitReturns(tagCommit, nil)

			actualReleaseDate, err := github.GetReleaseDate("0.11.0")
			require.NoError(err)

			assert.Equal("2020-12-31T00:00:00Z", actualReleaseDate.Format(time.RFC3339))
		})
	})
}
