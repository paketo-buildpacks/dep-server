package dependency_test

import (
	"testing"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPecl(t *testing.T) {
	spec.Run(t, "Pecl", testPecl, spec.Report(report.Terminal{}))
}

func testPecl(t *testing.T, when spec.G, it spec.S) {

	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		pecl                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		pecl, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("apc")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all source release final versions with newest versions first", func() {
			fakeWebClient.GetReturns([]byte(
				`<?xml version="1.0" encoding="utf-8"?>
<rdf:RDF
        xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
        xmlns="http://purl.org/rss/1.0/"
        xmlns:dc="http://purl.org/dc/elements/1.1/">

            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.13">
            <title>APC 3.1.13RC2</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.13</link>
            <dc:date>2012-09-03T13:15:17-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.12">
            <title>APC 3.1.12a1</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.12</link>
            <dc:date>2012-08-16T11:56:56-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.11">
            <title>APC 3.1.11b1</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.11</link>
            <dc:date>2012-07-19T18:11:06-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.10">
            <title>APC 3.1.10-beta1</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.10</link>
            <dc:date>2012-04-11T07:48:07-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.9">
            <title>APC 3.1.9</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.9</link>
            <dc:date>2011-05-14T18:15:04-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.8">
            <title>APC 3.1.8</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.8</link>
            <dc:date>2011-05-02T14:36:55-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.7">
            <title>APC 3.1.7</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.7</link>
            <dc:date>2011-01-11T14:10:06-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.6">
            <title>APC 3.1.6</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.6</link>
            <dc:date>2010-11-30T05:21:42-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.5">
            <title>APC 3.1.5</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.5</link>
            <dc:date>2010-11-02T14:21:27-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.4">
            <title>APC 3.1.4</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.4</link>
            <dc:date>2010-08-05T11:04:20-05:00</dc:date>
        </item>
</rdf:RDF>`), nil)
			versions, err := pecl.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"3.1.9",
				"3.1.8",
				"3.1.7",
				"3.1.6",
				"3.1.5",
				"3.1.4",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://pecl.php.net/feeds/pkg_apc.rss", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct version", func() {
			fakeWebClient.GetReturns([]byte(
				`<?xml version="1.0" encoding="utf-8"?>
<rdf:RDF
        xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
        xmlns="http://purl.org/rss/1.0/"
        xmlns:dc="http://purl.org/dc/elements/1.1/">

            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.8">
            <title>APC 3.1.8</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.8</link>
            <dc:date>2011-05-02T14:36:55-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.7">
            <title>APC 3.1.7</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.7</link>
            <dc:date>2011-01-11T14:10:06-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.6">
            <title>APC 3.1.6</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.6</link>
            <dc:date>2010-11-30T05:21:42-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.5">
            <title>APC 3.1.5</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.5</link>
            <dc:date>2010-11-02T14:21:27-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.4">
            <title>APC 3.1.4</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.4</link>
            <dc:date>2010-08-05T11:04:20-05:00</dc:date>
        </item>
</rdf:RDF>`), nil)
			fakeChecksummer.GetSHA256Returns("some-sha256", nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/pecl@3.1.6?checksum=some-sha256&download_url=https://pecl.php.net")

			actualDep, err := pecl.GetDependencyVersion("3.1.6")
			require.NoError(err)

			expectedReleaseDate, err := time.Parse(time.RFC3339, "2010-11-30T05:21:42-05:00")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedDep := dependency.DepVersion{
				Version:         "3.1.6",
				URI:             "https://pecl.php.net/get/APC-3.1.6",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				PURL:            "pkg:generic/pecl@3.1.6?checksum=some-sha256&download_url=https://pecl.php.net",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://pecl.php.net/feeds/pkg_apc.rss", urlArg)
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct release date", func() {
			fakeWebClient.GetReturns([]byte(
				`<?xml version="1.0" encoding="utf-8"?>
<rdf:RDF
        xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
        xmlns="http://purl.org/rss/1.0/"
        xmlns:dc="http://purl.org/dc/elements/1.1/">

            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.8">
            <title>APC 3.1.8</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.8</link>
            <dc:date>2011-05-02T14:36:55-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.7">
            <title>APC 3.1.7</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.7</link>
            <dc:date>2011-01-11T14:10:06-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.6">
            <title>APC 3.1.6</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.6</link>
            <dc:date>2010-11-30T05:21:42-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.5">
            <title>APC 3.1.5</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.5</link>
            <dc:date>2010-11-02T14:21:27-05:00</dc:date>
        </item>
            <item rdf:about="https://pecl.php.net/package-changelog.php?package=APC&amp;release=3.1.4">
            <title>APC 3.1.4</title>
            <link>https://pecl.php.net/package-changelog.php?package=APC&amp;amp;release=3.1.4</link>
            <dc:date>2010-08-05T11:04:20-05:00</dc:date>
        </item>
</rdf:RDF>`), nil)

			releaseDate, err := pecl.GetReleaseDate("3.1.6")
			require.NoError(err)

			assert.Equal("2010-11-30T05:21:42-05:00", releaseDate.Format(time.RFC3339))
		})
	})
}
