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

func TestRust(t *testing.T) {
	spec.Run(t, "Rust", testRust, spec.Report(report.Terminal{}))
}

func testRust(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		rust                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		rust, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("rust")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all rust release versions with newest versions first", func() {
			fakeGithubClient.GetTagsReturns([]string{
				"1.49.0",
				"1.48.0",
				"1.47.0",
				"1.46.0-alpha1",
				"0.2",
				"release-0.1",
			}, nil)

			versions, err := rust.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"1.49.0", "1.48.0", "1.47.0"}, versions)

			orgArg, repoArg := fakeGithubClient.GetTagsArgsForCall(0)
			assert.Equal("rust-lang", orgArg)
			assert.Equal("rust", repoArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct rust version", func() {
			date := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
			tagCommit := internal.GithubTagCommit{
				Tag:  "1.49.0",
				SHA:  "some-sha",
				Date: date,
			}
			fakeGithubClient.GetTagCommitReturns(tagCommit, nil)
			fakeWebClient.GetReturnsOnCall(0, []byte("some-gpg-key"), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte("some-signature"), nil)

			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/rust@1.49.0?checksum=some-source-sha&download_url=https://static.rust-lang.org")

			actualDepVersion, err := rust.GetDependencyVersion("1.49.0")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedDepVersion := dependency.DepVersion{
				Version:         "1.49.0",
				URI:             "https://static.rust-lang.org/dist/rustc-1.49.0-src.tar.gz",
				SHA256:          "some-source-sha",
				ReleaseDate:     &tagCommit.Date,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:rust-lang:rust:1.49.0:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/rust@1.49.0?checksum=some-source-sha&download_url=https://static.rust-lang.org",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDepVersion, actualDepVersion)

			url, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://static.rust-lang.org/rust-key.gpg.ascii", url)

			url, _ = fakeWebClient.GetArgsForCall(1)
			assert.Equal("https://static.rust-lang.org/dist/rustc-1.49.0-src.tar.gz.asc", url)

			urlArg, _, _ := fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("https://static.rust-lang.org/dist/rustc-1.49.0-src.tar.gz", urlArg)

			releaseAssetSignatureArg, _, rustGPGKeyArg := fakeChecksummer.VerifyASCArgsForCall(0)
			assert.Equal("some-signature", releaseAssetSignatureArg)
			assert.Equal([]string{"some-gpg-key"}, rustGPGKeyArg)
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct rust release date", func() {
			date := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
			tagCommit := internal.GithubTagCommit{
				Tag:  "1.49.0",
				SHA:  "some-sha",
				Date: date,
			}
			fakeGithubClient.GetTagCommitReturns(tagCommit, nil)

			actualReleaseDate, err := rust.GetReleaseDate("1.49.0")
			require.NoError(err)

			assert.Equal("2020-12-31T00:00:00Z", actualReleaseDate.Format(time.RFC3339))
		})
	})
}
