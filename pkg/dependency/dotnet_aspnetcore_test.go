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
)

func TestDotnetASPNETCore(t *testing.T) {
	spec.Run(t, "Dotnet ASP.NET Core", testDotnetASPNETCore, spec.Report(report.Terminal{}))
}

func testDotnetASPNETCore(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		dotnetASPNETCore     dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		dotnetASPNETCore, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("dotnet-aspnetcore")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all dotnet aspnetcore final versions with newest versions first", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{"releases-index": [{"channel-version": "2.0"}, {"channel-version": "1.1"}, {"channel-version": "1.0"}]}
`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {"version": "2.0.2"}
    },
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": null
    },
    {
      "release-date": "2020-02-19",
      "aspnetcore-runtime": null
    },
    {
      "release-date": "2020-02-10",
      "aspnetcore-runtime": {"version": "2.0.1"}
    },
    {
      "release-date": "2020-02-01",
      "aspnetcore-runtime": {"version": "2.0.0"}
    },
    {
      "release-date": "2020-02-01",
      "aspnetcore-runtime": {"version": "2.0.0-preview1.20000.20"}
    },
    {
      "release-date": "2020-02-01",
      "aspnetcore-runtime": {"version": "2.0.0-preview1-20000-20"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {"version": "1.1.2"}
    },
    {
      "release-date": "2020-02-11",
      "aspnetcore-runtime": {"version": "1.1.1"}
    },
    {
      "release-date": "2020-02-01",
      "aspnetcore-runtime": {"version": "1.1.1"}
    },
    {
      "release-date": "2020-01-15",
      "aspnetcore-runtime": {"version": "1.1.0"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(3, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-05",
      "aspnetcore-runtime": {"version": "1.0.2"}
    },
    {
      "release-date": "2020-02-01",
      "aspnetcore-runtime": {"version": "1.0.1"}
    },
    {
      "release-date": "2020-01-10",
      "aspnetcore-runtime": {"version": "1.0.0"}
    }
  ]
}
`), nil)

			versions, err := dotnetASPNETCore.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"2.0.2",
				"1.1.2",
				"2.0.1",
				"1.0.2",
				"2.0.0",
				"1.1.1",
				"1.0.1",
				"1.1.0",
				"1.0.0",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/releases-index.json", urlArg)
			urlArg, _ = fakeWebClient.GetArgsForCall(1)
			assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/2.0/releases.json", urlArg)
			urlArg, _ = fakeWebClient.GetArgsForCall(2)
			assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/1.1/releases.json", urlArg)
			urlArg, _ = fakeWebClient.GetArgsForCall(3)
			assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/1.0/releases.json", urlArg)
		})

		it("deduplicates versions across channels", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{"releases-index": [{"channel-version": "2.0"}, {"channel-version": "1.1"}]}
`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {"version": "2.0.1"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-15",
      "aspnetcore-runtime": {"version": "1.1.2"}
    },
    {
      "release-date": "2020-02-10",
      "aspnetcore-runtime": {"version": "2.0.1"}
    }
  ]
}
`), nil)

			versions, err := dotnetASPNETCore.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"1.1.2",
				"2.0.1",
			}, versions)
		})
	})

	when("GetDependencyVersion", func() {
		var (
			expectedReleaseDate     = time.Date(2020, 02, 20, 0, 0, 0, 0, time.UTC)
			expectedDeprecationDate = time.Date(2050, 02, 20, 0, 0, 0, 0, time.UTC)
		)

		it("returns the correct dotnet aspnetcore version", func() {
			fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "aspnetcore-runtime": {
        "version": "2.0.2",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.2",
            "hash": "sha512-for-linux-x64-2.0.2"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.1",
            "hash": "SHA512-FOR-LINUX-X64-2.0.1"
          },
          {
            "name": "aspnetcore-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.1",
            "hash": "sha512-for-osx-64-2.0.1"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-10",
      "aspnetcore-runtime": {
        "version": "2.0.0",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.0",
            "hash": "sha512-for-linux-x64-2.0.0"
          }
        ]
      }
    }
  ]
}
`), nil)
			fakeChecksummer.GetSHA256Returns("some-sha256", nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

			actualDep, err := dotnetASPNETCore.GetDependencyVersion("2.0.1")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedDep := dependency.DepVersion{
				Version:         "2.0.1",
				URI:             "url-for-linux-x64-2.0.1",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: &expectedDeprecationDate,
				CPE:             "cpe:2.3:a:microsoft:asp.net_core:2.0:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/2.0/releases.json", urlArg)

			urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("url-for-linux-x64-2.0.1", urlArg)

			_, sha512Arg := fakeChecksummer.VerifySHA512ArgsForCall(0)
			assert.Equal("sha512-for-linux-x64-2.0.1", sha512Arg)
		})

		when("the file rid is ubuntu-x64", func() {
			it("returns the correct dotnet aspnetcore version", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "aspnetcore-runtime-ubuntu-x64.tar.gz",
            "rid": "ubuntu-x64",
            "url": "url-for-ubuntu-x64-2.0.1",
            "hash": "SHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA512"
          },
          {
            "name": "aspnetcore-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.1",
            "hash": "sha512-for-osx-64-2.0.1"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetASPNETCore.GetDependencyVersion("2.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())

				expectedDep := dependency.DepVersion{
					Version:         "2.0.1",
					URI:             "url-for-ubuntu-x64-2.0.1",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:asp.net_core:2.0:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/2.0/releases.json", urlArg)

				urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("url-for-ubuntu-x64-2.0.1", urlArg)

				_, sha512Arg := fakeChecksummer.VerifySHA512ArgsForCall(0)
				assert.Equal("shaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa512", sha512Arg)
			})
		})

		when("there is no linux or ubuntu file in the release", func() {
			it("returns a NoSourceCodeError", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "aspnetcore-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.1",
            "hash": "sha512-for-osx-64-2.0.1"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)

				_, err := dotnetASPNETCore.GetDependencyVersion("2.0.1")
				assert.Error(err)

				assert.True(errors.Is(err, depErrors.NoSourceCodeError{Version: "2.0.1"}))
			})
		})

		when("the file hash is empty", func() {
			it("does not try to verify the hash", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.1",
            "hash": ""
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetASPNETCore.GetDependencyVersion("2.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())

				expectedDep := dependency.DepVersion{
					Version:         "2.0.1",
					URI:             "url-for-linux-x64-2.0.1",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:asp.net_core:2.0:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				assert.Equal(0, fakeChecksummer.VerifySHA512CallCount())
			})
		})

		when("the file has is a sha256", func() {
			it("returns the hash and does not download the file", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.1",
            "hash": "SHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA256"
          }
        ]
      }
    }
  ]
}
`), nil)

				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetASPNETCore.GetDependencyVersion("2.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())

				expectedDep := dependency.DepVersion{
					Version:         "2.0.1",
					URI:             "url-for-linux-x64-2.0.1",
					SHA256:          "shaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:asp.net_core:2.0:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				assert.Equal(0, fakeWebClient.DownloadCallCount())
				assert.Equal(0, fakeChecksummer.VerifySHA512CallCount())
			})
		})

		when("the channel's eol date is empty", func() {
			it("returns an empty eol date", func() {

				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "",
  "releases": [
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.2",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.2",
            "hash": "SHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA512"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetASPNETCore.GetDependencyVersion("2.0.2")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())

				expectedDep := dependency.DepVersion{
					Version:         "2.0.2",
					URI:             "url-for-linux-x64-2.0.2",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:microsoft:asp.net_core:2.0:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-aspnetcore@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct dotnet aspnetcore release date", func() {
			fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "aspnetcore-runtime": {
        "version": "2.0.2",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.2",
            "hash": "sha512-for-linux-x64-2.0.2"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-20",
      "aspnetcore-runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.1",
            "hash": "SHA512-FOR-LINUX-X64-2.0.1"
          },
          {
            "name": "aspnetcore-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.1",
            "hash": "sha512-for-osx-64-2.0.1"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-10",
      "aspnetcore-runtime": {
        "version": "2.0.0",
        "files": [
          {
            "name": "aspnetcore-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.0",
            "hash": "sha512-for-linux-x64-2.0.0"
          }
        ]
      }
    }
  ]
}
`), nil)
			fakeChecksummer.GetSHA256Returns("some-sha256", nil)

			releaseDate, err := dotnetASPNETCore.GetReleaseDate("2.0.1")
			require.NoError(err)

			assert.Equal("2020-02-20T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
