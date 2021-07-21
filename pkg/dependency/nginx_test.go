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

func TestNginx(t *testing.T) {
	spec.Run(t, "Nginx", testNginx, spec.Report(report.Terminal{}))
}

func testNginx(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		nginx                dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		nginx, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("nginx")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all nginx release versions with newest versions first", func() {
			fakeGithubClient.GetTagsReturns([]string{
				"release-3.0.0",
				"release-1.0.1",
				"release-2.0.0",
				"release-1.0.0",
			}, nil)

			versions, err := nginx.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"3.0.0", "1.0.1", "2.0.0", "1.0.0"}, versions)

			orgArg, repoArg := fakeGithubClient.GetTagsArgsForCall(0)
			assert.Equal("nginx", orgArg)
			assert.Equal("nginx", repoArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct nginx version", func() {
			fakeGithubClient.GetTagCommitReturns(internal.GithubTagCommit{
				Tag:  "release-1.0.0",
				SHA:  "dddddddddddddddddddddddddddddddddddddddd",
				Date: time.Date(2020, 06, 17, 0, 0, 0, 0, time.UTC),
			}, nil)
			fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte("some-signature"), nil)
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/nginx@1.0.0?checksum=some-source-sha&download_url=http://nginx.org")

			actualDepVersion, err := nginx.GetDependencyVersion("1.0.0")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 06, 17, 0, 0, 0, 0, time.UTC)
			expectedDepVersion := dependency.DepVersion{
				Version:         "1.0.0",
				URI:             "http://nginx.org/download/nginx-1.0.0.tar.gz",
				SHA256:          "some-source-sha",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:nginx:nginx:1.0.0:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/nginx@1.0.0?checksum=some-source-sha&download_url=http://nginx.org",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDepVersion, actualDepVersion)

			url, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("http://nginx.org/keys/mdounin.key", url)

			urlArg, _, _ := fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("http://nginx.org/download/nginx-1.0.0.tar.gz", urlArg)

			releaseAssetSignatureArg, _, nginxGPGKeyArg := fakeChecksummer.VerifyASCArgsForCall(0)
			assert.Equal("some-signature", releaseAssetSignatureArg)
			assert.Equal([]string{"some-gpg-key"}, nginxGPGKeyArg)
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct nginx release date", func() {
			fakeGithubClient.GetTagCommitReturns(internal.GithubTagCommit{
				Tag:  "release-1.0.0",
				SHA:  "dddddddddddddddddddddddddddddddddddddddd",
				Date: time.Date(2020, 06, 17, 0, 0, 0, 0, time.UTC),
			}, nil)

			releaseDate, err := nginx.GetReleaseDate("1.0.0")
			require.NoError(err)

			assert.Equal("2020-06-17T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
