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
)

func TestRuby(t *testing.T) {
	spec.Run(t, "Ruby", testRuby, spec.Report(report.Terminal{}))
}

func testRuby(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		ruby                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		ruby, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("ruby")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all ruby release versions with newest versions first", func() {
			fakeWebClient.GetReturns([]byte(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">

<h3>Ruby releases by version number</h3>


<table class="release-list">
<tr>
<th>Release Version</th>
<th>Release Date</th>
<th>Release Notes</th>
</tr>


<tr>
<td>Ruby 2.7.1</td>
<td>2020-04-31</td>
<td><a href="/en/news/2020/03/31/ruby-2-7-1-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.6.6</td>
<td>2020-04-31</td>
<td><a href="/en/news/2020/03/31/ruby-2-6-6-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.5.8</td>
<td>2020-02-31</td>
<td><a href="/en/news/2020/03/31/ruby-2-5-8-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.4.10</td>
<td>2020-01-31</td>
<td><a href="/en/news/2020/03/31/ruby-2-4-10-released/">more...</a></td>
</tr>
</tr>
</table>

  </div>
</div>
<hr class="hidden-modern" />

<div id="sidebar-wrapper">
  <div id="sidebar">
  </body>
</html>

`), nil)

			versions, err := ruby.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"2.7.1", "2.6.6", "2.5.8", "2.4.10"}, versions)
		})
	})

	when("GetDependencyVersion", func() {
		when("the version IS in the release yaml file", func() {
			it("returns the correct ruby version", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
<tr>
<td>Ruby 3.0.0</td>
<td>2020-12-25</td>
<td><a href="/en/news/2020/12/25/ruby-3-0-0-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.7.1</td>
<td>2020-03-20</td>
<td><a href="/en/news/2020/03/31/ruby-2-7-1-released/">more...</a></td>
</tr>
</table>
`), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte(`
- version: 3.0.0
  url:
    gz:  https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.gz
    zip: https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.zip
    xz:  https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.xz
  sha256:
    gz:  some-sha-256-gz
    zip: some-sha-256-zip
    xz:  some-sha-256-xz

- version: 2.7.2
  url:
    bz2: https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.2.tar.bz2
    gz: https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.2.tar.gz
    xz: https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.2.tar.xz
    zip: https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.2.zip
  sha256:
    bz2: some-other-sha-256-bz2
    gz: some-other-sha-256-gz
    xz: some-other-sha-256-xz
    zip: some-other-sha-256-zip
`), nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/ruby@3.0.0?checksum=some-sha-256-gz&download_url=https://cache.ruby-lang.org")

				actualDepVersion, err := ruby.GetDependencyVersion("3.0.0")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedReleaseDate := time.Date(2020, 12, 25, 0, 0, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "3.0.0",
					URI:             "https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.gz",
					SHA256:          "some-sha-256-gz",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:ruby-lang:ruby:3.0.0:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/ruby@3.0.0?checksum=some-sha-256-gz&download_url=https://cache.ruby-lang.org",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDepVersion, actualDepVersion)

				url, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://www.ruby-lang.org/en/downloads/releases/", url)

				url, _ = fakeWebClient.GetArgsForCall(1)
				assert.Equal("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml", url)
			})
		})

		when("the version IS in the release yaml file but it does NOT have the URL and SHA256", func() {
			it("check the index.txt file", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
<tr>
<td>Ruby 1.6.7</td>
<td>2002-03-01</td>
<td><a href="/en/news/2002/03/01/ruby-1-6-7-released/">more...</a></td>
</tr>
`), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte(`
- version: 1.6.7
  date: 2002-03-01
`), nil)

				fakeWebClient.GetReturnsOnCall(2, []byte(`
ruby-1.6.7	https://cache.ruby-lang.org/pub/ruby/1.6/ruby-1.6.7.tar.gz	some-sha-1	some-sha-256	some-sha-512
`), nil)

				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

				actualDepVersion, err := ruby.GetDependencyVersion("1.6.7")
				require.NoError(err)

				expectedReleaseDate := time.Date(2002, 3, 1, 0, 0, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "1.6.7",
					URI:             "https://cache.ruby-lang.org/pub/ruby/1.6/ruby-1.6.7.tar.gz",
					SHA256:          "some-sha-256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:ruby-lang:ruby:1.6.7:*:*:*:*:*:*:*",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDepVersion, actualDepVersion)

				url, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://www.ruby-lang.org/en/downloads/releases/", url)

				url, _ = fakeWebClient.GetArgsForCall(1)
				assert.Equal("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml", url)

				url, _ = fakeWebClient.GetArgsForCall(2)
				assert.Equal("https://cache.ruby-lang.org/pub/ruby/index.txt", url)
			})
		})

		when("the version is NOT in the release yaml file but IS in index.txt", func() {
			it("returns the correct ruby version", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
<tr>
<td>Ruby 2.7.1</td>
<td>2020-03-20</td>
<td><a href="/en/news/2020/03/31/ruby-2-7-1-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.6.6</td>
<td>2020-04-20</td>
<td><a href="/en/news/2020/03/31/ruby-2-6-6-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.5.8</td>
<td>2020-02-20</td>
<td><a href="/en/news/2020/03/31/ruby-2-5-8-released/">more...</a></td>
</tr>
`), nil)

				fakeWebClient.GetReturnsOnCall(1, []byte(`
- version: 3.0.0
  url:
    gz:  https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.gz
    zip: https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.zip
    xz:  https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.xz
  sha256:
    gz:  some-wrong-sha-256-gz
    zip: some-wrong-sha-256-zip
    xz:  some-wrong-sha-256-xz
`), nil)

				fakeWebClient.GetReturnsOnCall(2, []byte(`
ruby-2.6.6	https://cache.ruby-lang.org/pub/ruby/2.6/ruby-2.6.6.zip	some-sha-1	some-sha-256-zip	some-sha-512
ruby-2.6.6	https://cache.ruby-lang.org/pub/ruby/2.6/ruby-2.6.6.tar.gz	some-sha-1	some-sha-256-gz	some-sha-512
ruby-2.7.1	https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.1.tar.bz	some-sha-1	some-other-sha-256-bz	some-sha-512
ruby-2.7.1	https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.1.tar.gz	some-sha-1	some-other-sha-256-gz	some-sha-512
`), nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

				actualDepVersion, err := ruby.GetDependencyVersion("2.6.6")
				require.NoError(err)

				expectedReleaseDate := time.Date(2020, 4, 20, 0, 0, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "2.6.6",
					URI:             "https://cache.ruby-lang.org/pub/ruby/2.6/ruby-2.6.6.tar.gz",
					SHA256:          "some-sha-256-gz",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:ruby-lang:ruby:2.6.6:*:*:*:*:*:*:*",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDepVersion, actualDepVersion)

				url, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://www.ruby-lang.org/en/downloads/releases/", url)

				url, _ = fakeWebClient.GetArgsForCall(1)
				assert.Equal("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml", url)

				url, _ = fakeWebClient.GetArgsForCall(2)
				assert.Equal("https://cache.ruby-lang.org/pub/ruby/index.txt", url)
			})

			when("the version on the download page ends in -0", func() {
				it("returns the correct ruby version", func() {
					fakeWebClient.GetReturnsOnCall(0, []byte(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">

<h3>Ruby releases by version number</h3>


<table class="release-list">
<tr>
<th>Release Version</th>
<th>Release Date</th>
<th>Release Notes</th>
</tr>


<tr>
<td>Ruby 1.9.0</td>
<td>2007-12-25</td>
<td><a href="/en/news/2007/12/25/ruby-1-9-0-released/">more...</a></td>
</tr>
</table>

  </div>
</div>
<hr class="hidden-modern" />

<div id="sidebar-wrapper">
  <div id="sidebar">
  </body>
</html>

`), nil)

					fakeWebClient.GetReturnsOnCall(1, []byte(`
- version: 3.0.0
  url:
    gz:  https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.gz
  sha256:
    gz:  some-wrong-sha-256-gz
`), nil)

					fakeWebClient.GetReturnsOnCall(2, []byte(`
ruby-1.9.0-0	https://cache.ruby-lang.org/pub/ruby/1.9/ruby-1.9.0-0.tar.gz	some-sha-1	some-sha-256	some-sha-512
`), nil)
					fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

					actualDepVersion, err := ruby.GetDependencyVersion("1.9.0")
					require.NoError(err)

					expectedReleaseDate := time.Date(2007, 12, 25, 0, 0, 0, 0, time.UTC)
					expectedDepVersion := dependency.DepVersion{
						Version:         "1.9.0",
						URI:             "https://cache.ruby-lang.org/pub/ruby/1.9/ruby-1.9.0-0.tar.gz",
						SHA256:          "some-sha-256",
						ReleaseDate:     &expectedReleaseDate,
						DeprecationDate: nil,
						CPE:             "cpe:2.3:a:ruby-lang:ruby:1.9.0:*:*:*:*:*:*:*",
						Licenses:        []string{"MIT", "MIT-2"},
					}
					assert.Equal(expectedDepVersion, actualDepVersion)

					url, _ := fakeWebClient.GetArgsForCall(0)
					assert.Equal("https://www.ruby-lang.org/en/downloads/releases/", url)

					url, _ = fakeWebClient.GetArgsForCall(1)
					assert.Equal("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml", url)

					url, _ = fakeWebClient.GetArgsForCall(2)
					assert.Equal("https://cache.ruby-lang.org/pub/ruby/index.txt", url)
				})
			})

			when("the version on the download page ends in -p0", func() {
				it("returns the correct ruby version", func() {
					fakeWebClient.GetReturnsOnCall(0, []byte(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">

<h3>Ruby releases by version number</h3>


<table class="release-list">
<tr>
<th>Release Version</th>
<th>Release Date</th>
<th>Release Notes</th>
</tr>


<tr>
<td>Ruby 1.9.1</td>
<td>2009-01-30</td>
<td><a href="/en/news/2009/01/30/ruby-1-9-1-released/">more...</a></td>
</tr>
</table>

  </div>
</div>
<hr class="hidden-modern" />

<div id="sidebar-wrapper">
  <div id="sidebar">
  </body>
</html>

`), nil)

					fakeWebClient.GetReturnsOnCall(1, []byte(`
- version: 3.0.0
  url:
    gz:  https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.gz
  sha256:
    gz:  some-wrong-sha-256-gz
`), nil)

					fakeWebClient.GetReturnsOnCall(2, []byte(`
ruby-1.9.1-p0	https://cache.ruby-lang.org/pub/ruby/1.9/ruby-1.9.1-p0.tar.gz	some-sha-1	some-sha-256	some-sha-512
`), nil)
					fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

					actualDepVersion, err := ruby.GetDependencyVersion("1.9.1")
					require.NoError(err)

					expectedReleaseDate := time.Date(2009, 1, 30, 0, 0, 0, 0, time.UTC)
					expectedDepVersion := dependency.DepVersion{
						Version:         "1.9.1",
						URI:             "https://cache.ruby-lang.org/pub/ruby/1.9/ruby-1.9.1-p0.tar.gz",
						SHA256:          "some-sha-256",
						ReleaseDate:     &expectedReleaseDate,
						DeprecationDate: nil,
						CPE:             "cpe:2.3:a:ruby-lang:ruby:1.9.1:*:*:*:*:*:*:*",
						Licenses:        []string{"MIT", "MIT-2"},
					}
					assert.Equal(expectedDepVersion, actualDepVersion)

					url, _ := fakeWebClient.GetArgsForCall(0)
					assert.Equal("https://www.ruby-lang.org/en/downloads/releases/", url)

					url, _ = fakeWebClient.GetArgsForCall(1)
					assert.Equal("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml", url)

					url, _ = fakeWebClient.GetArgsForCall(2)
					assert.Equal("https://cache.ruby-lang.org/pub/ruby/index.txt", url)
				})
			})
		})
	})

	when("GetReleaseDate", func() {
		when("the version IS in the release yaml file", func() {
			it("returns the correct ruby release date", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
<tr>
<td>Ruby 3.0.0</td>
<td>2020-12-25</td>
<td><a href="/en/news/2020/12/25/ruby-3-0-0-released/">more...</a></td>
</tr>
<tr>
<td>Ruby 2.7.1</td>
<td>2020-03-20</td>
<td><a href="/en/news/2020/03/31/ruby-2-7-1-released/">more...</a></td>
</tr>
</table>
`), nil)

				releaseDate, err := ruby.GetReleaseDate("3.0.0")
				require.NoError(err)

				assert.Equal("2020-12-25T00:00:00Z", releaseDate.Format(time.RFC3339))
			})
		})
	})
}
