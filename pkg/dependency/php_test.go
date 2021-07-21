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

func TestPhp(t *testing.T) {
	spec.Run(t, "Php", testPhp, spec.Report(report.Terminal{}))
}

func testPhp(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		php                  dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		php, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("php")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all php release versions with newest versions first", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{
  "5": {},
  "7": {}
}`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "7.4.4": {"date": "21 Mar 2020", "source": [{"filename": "php-7.4.4.tar.gz", "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]},
  "7.3.16": {"date": "21 Mar 2019", "source": [{"filename": "php-7.3.16.tar.gz", "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]},
  "7.2.29": {"date": "21 Mar 2018", "source": [{"filename": "php-7.2.29.tar.gz", "sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}]}
}`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "5.6.40": {"date": "20 Mar 2020", "source": [{"filename": "php-5.6.40.tar.gz", "sha256": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}]},
  "5.6.39": {"date": "20 Mar 2019", "source": [{"filename": "php-5.6.39.tar.gz", "sha256": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"}]},
  "5.6.38": {"date": "20 Mar 2018", "source": [{"filename": "php-5.6.38.tar.gz", "sha256": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}]}
}`), nil)
			versions, err := php.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"7.4.4", "5.6.40", "7.3.16", "5.6.39", "7.2.29", "5.6.38"}, versions)

			allVersionsUrlArg, _ := fakeWebClient.GetArgsForCall(0)
			version5UrlArg, _ := fakeWebClient.GetArgsForCall(1)
			version7UrlArg, _ := fakeWebClient.GetArgsForCall(2)

			assert.Equal("https://www.php.net/releases/index.php?json", allVersionsUrlArg)
			assert.Equal("https://www.php.net/releases/index.php?json&version=5&max=1000", version5UrlArg)
			assert.Equal("https://www.php.net/releases/index.php?json&version=7&max=1000", version7UrlArg)
		})

		it("does not skip versions with wrong date scheme", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{
  "7": {}
}`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "7.4.4": {"date": "21 Mar 2020", "source": [{"filename": "php-7.4.4.tar.gz", "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]},
  "7.3.16": {"date": "21 Mar 2019", "source": [{"filename": "php-7.3.16.tar.gz", "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]},
  "7.2.29": {"date": "2 Mar 2018", "source": [{"filename": "php-7.2.29.tar.gz", "sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}]},
  "7.2.28": {"date": "02 March 2018", "source": [{"filename": "php-7.2.28.tar.gz", "sha256": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}]},
  "7.2.27": {"date": "2 March 2018", "source": [{"filename": "php-7.2.27.tar.gz", "sha256": "eeeeeeeeeeccccccccccccccccccccccccccccccccccccccccccccccccccccce"}]}
}`), nil)
			versions, err := php.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"7.4.4", "7.3.16", "7.2.29", "7.2.28", "7.2.27"}, versions)
		})

		it("sorts versions released on the same day in numerical order", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{
  "7": {}
}`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "7.4.4": {"date": "19 Mar 2020", "source": [{"filename": "php-7.4.4.tar.gz", "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}]},
  "7.2.29": {"date": "19 Mar 2020", "source": [{"filename": "php-7.2.29.tar.gz", "sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}]},
  "7.3.16": {"date": "19 Mar 2020", "source": [{"filename": "php-7.3.16.tar.gz", "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]}
}`), nil)
			versions, err := php.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"7.4.4", "7.3.16", "7.2.29"}, versions)
		})

		it("does not include versions from the 3.x line", func() {
			// The version 3 index does not have proper versions in them, it has a single version which is '3.0.x (latest)'

			fakeWebClient.GetReturnsOnCall(0, []byte(`
{
  "3": {},
  "5": {}
}`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "5.6.40": {"date": "20 Mar 2020", "source": [{"filename": "php-5.6.40.tar.gz", "sha256": "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}]}
}`), nil)
			versions, err := php.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"5.6.40"}, versions)
			assert.Equal(2, fakeWebClient.GetCallCount())
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct php version", func() {
			fakeWebClient.GetReturns([]byte(`
{
 "date": "19 Mar 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "19 Mar 2020"
  },
  {
   "filename": "php-7.4.4.tar.bz",
   "name": "PHP 7.4.4 (tar.bz)",
   "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
   "date": "19 Mar 2020"
  }
 ]
}
`), nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/php@7.4.4?checksum=aaaaaa&download_url=https://www.php.net")

			actualDepVersion, err := php.GetDependencyVersion("7.4.4")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 03, 19, 0, 0, 0, 0, time.UTC)
			expectedDeprecationDate := time.Date(2023, 03, 19, 0, 0, 0, 0, time.UTC)
			expectedDepVersion := dependency.DepVersion{
				Version:         "7.4.4",
				URI:             "https://www.php.net/distributions/php-7.4.4.tar.gz",
				SHA256:          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: &expectedDeprecationDate,
				CPE:             "cpe:2.3:a:php:php:7.4.4:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/php@7.4.4?checksum=aaaaaa&download_url=https://www.php.net",
				Licenses:        []string{"MIT", "MIT-2"},
			}

			assert.Equal(expectedDepVersion, actualDepVersion)

			url, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://www.php.net/releases/index.php?json&version=7.4.4", url)
		})

		when("the dependency has an MD5 instead of a SHA256", func() {
			it("validates the MD5 and calculates the SHA256", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "19 Mar 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "md5": "some-md5",
   "date": "19 Mar 2020"
  }
 ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

				actualDepVersion, err := php.GetDependencyVersion("7.4.4")
				require.NoError(err)

				expectedReleaseDate := time.Date(2020, 03, 19, 0, 0, 0, 0, time.UTC)
				expectedDeprecationDate := time.Date(2023, 03, 19, 0, 0, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "7.4.4",
					URI:             "https://www.php.net/distributions/php-7.4.4.tar.gz",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:php:php:7.4.4:*:*:*:*:*:*:*",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDepVersion, actualDepVersion)

				url, _, _ := fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("https://www.php.net/distributions/php-7.4.4.tar.gz", url)

				_, md5Arg := fakeChecksummer.VerifyMD5ArgsForCall(0)
				assert.Equal("some-md5", md5Arg)
			})
		})

		when("the dependency version is known to have an incorrect checksum", func() {
			it("does not try to validate it", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "09 May 2013",
 "source": [
  {
   "filename": "php-5.3.25.tar.gz",
   "name": "PHP 5.3.25 (tar.gz)",
   "md5": "some-incorrect-md5",
   "date": "09 May 2013"
  }
 ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

				actualDepVersion, err := php.GetDependencyVersion("5.3.25")
				require.NoError(err)

				expectedReleaseDate := time.Date(2013, 05, 9, 0, 0, 0, 0, time.UTC)
				expectedDeprecationDate := time.Date(2016, 05, 9, 0, 0, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "5.3.25",
					URI:             "https://www.php.net/distributions/php-5.3.25.tar.gz",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:php:php:5.3.25:*:*:*:*:*:*:*",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDepVersion, actualDepVersion)

				assert.Equal(0, fakeChecksummer.VerifyMD5CallCount())
			})
		})

		when("the dependency version is known to be missing a checksum", func() {
			it("does not try to validate it", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "24 Aug 2006",
 "source": [
  {
   "filename": "php-5.1.6.tar.gz",
   "name": "PHP 5.1.6 (tar.gz)"
  }
 ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)

				actualDepVersion, err := php.GetDependencyVersion("5.1.6")
				require.NoError(err)

				expectedReleaseDate := time.Date(2006, 8, 24, 0, 0, 0, 0, time.UTC)
				expectedDeprecationDate := time.Date(2009, 8, 24, 0, 0, 0, 0, time.UTC)
				expectedDepVersion := dependency.DepVersion{
					Version:         "5.1.6",
					URI:             "https://www.php.net/distributions/php-5.1.6.tar.gz",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:php:php:5.1.6:*:*:*:*:*:*:*",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDepVersion, actualDepVersion)

				assert.Equal(0, fakeChecksummer.VerifyMD5CallCount())
			})
		})

		when("the download is in the museum", func() {
			it("returns the museum URL", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "08 Feb 2007",
 "source": [
  {
   "filename": "php-5.2.1.tar.gz",
   "name": "PHP 5.2.1 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "08 Feb 2007"
  }
 ],
 "museum": true
}
`), nil)

				depVersion, err := php.GetDependencyVersion("5.2.1")
				require.NoError(err)

				assert.Equal("https://museum.php.net/php5/php-5.2.1.tar.gz", depVersion.URI)
			})
		})

		when("the release date has an unpadded day", func() {
			it("properly parses the date", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "1 Mar 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "1 Mar 2020"
  }
 ]
}
`), nil)

				depVersion, err := php.GetDependencyVersion("7.4.4")
				require.NoError(err)

				assert.Equal("2020-03-01T00:00:00Z", depVersion.ReleaseDate.Format(time.RFC3339))
			})
		})

		when("the release date has a full month name", func() {
			it("properly parses the date", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "01 March 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "01 March 2020"
  }
 ]
}
`), nil)

				depVersion, err := php.GetDependencyVersion("7.4.4")
				require.NoError(err)

				assert.Equal("2020-03-01T00:00:00Z", depVersion.ReleaseDate.Format(time.RFC3339))
			})
		})

		when("the release date has an unpadded day and a full month name", func() {
			it("properly parses the date", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "1 March 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "1 March 2020"
  }
 ]
}
`), nil)

				depVersion, err := php.GetDependencyVersion("7.4.4")
				require.NoError(err)

				assert.Equal("2020-03-01T00:00:00Z", depVersion.ReleaseDate.Format(time.RFC3339))
			})
		})

		when("the source cannot be found", func() {
			it("returns an error", func() {
				fakeWebClient.GetReturns([]byte(`{"error": "Unknown version"}`), nil)
				_, err := php.GetDependencyVersion("7.4.4")
				assert.Error(err)
				assert.Equal("could not find .tar.gz file for 7.4.4", err.Error())
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct php release date", func() {
			fakeWebClient.GetReturns([]byte(`
{
 "date": "19 Mar 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "19 Mar 2020"
  },
  {
   "filename": "php-7.4.4.tar.bz",
   "name": "PHP 7.4.4 (tar.bz)",
   "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
   "date": "19 Mar 2020"
  }
 ]
}
`), nil)

			releaseDate, err := php.GetReleaseDate("7.4.4")
			require.NoError(err)

			assert.Equal("2020-03-19T00:00:00Z", releaseDate.Format(time.RFC3339))
		})

		when("the release date has an unpadded day", func() {
			it("properly parses the date", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "1 Mar 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "1 Mar 2020"
  }
 ]
}
`), nil)

				releaseDate, err := php.GetReleaseDate("7.4.4")
				require.NoError(err)

				assert.Equal("2020-03-01T00:00:00Z", releaseDate.Format(time.RFC3339))
			})
		})

		when("the release date has a full month name", func() {
			it("properly parses the date", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "01 March 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "01 March 2020"
  }
 ]
}
`), nil)

				releaseDate, err := php.GetReleaseDate("7.4.4")
				require.NoError(err)

				assert.Equal("2020-03-01T00:00:00Z", releaseDate.Format(time.RFC3339))
			})
		})

		when("the release date has an unpadded day and a full month name", func() {
			it("properly parses the date", func() {
				fakeWebClient.GetReturns([]byte(`
{
 "date": "1 March 2020",
 "source": [
  {
   "filename": "php-7.4.4.tar.gz",
   "name": "PHP 7.4.4 (tar.gz)",
   "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
   "date": "1 March 2020"
  }
 ]
}
`), nil)

				releaseDate, err := php.GetReleaseDate("7.4.4")
				require.NoError(err)

				assert.Equal("2020-03-01T00:00:00Z", releaseDate.Format(time.RFC3339))
			})
		})
	})
}
