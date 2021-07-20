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

func TestDotnetRuntime(t *testing.T) {
	spec.Run(t, "Dotnet runtime", testDotnetRuntime, spec.Report(report.Terminal{}))
}

func testDotnetRuntime(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		dotnetRuntime        dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		dotnetRuntime, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("dotnet-runtime")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all dotnet runtime final versions with newest versions first", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{"releases-index": [{"channel-version": "2.0"}, {"channel-version": "1.1"}, {"channel-version": "1.0"}]}
`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "runtime": {"version": "2.0.2"}
    },
    {
      "release-date": "2020-02-20",
      "runtime": null
    },
    {
      "release-date": "2020-02-19",
      "runtime": null
    },
    {
      "release-date": "2020-02-10",
      "runtime": {"version": "2.0.1"}
    },
    {
      "release-date": "2020-02-01",
      "runtime": {"version": "2.0.0"}
    },
    {
      "release-date": "2020-02-01",
      "runtime": {"version": "2.0.0-preview1.20000.20"}
    },
    {
      "release-date": "2020-02-01",
      "runtime": {"version": "2.0.0-preview1-20000-20"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "runtime": {"version": "1.1.2"}
    },
    {
      "release-date": "2020-02-11",
      "runtime": {"version": "1.1.1"}
    },
    {
      "release-date": "2020-02-01",
      "runtime": {"version": "1.1.1"}
    },
    {
      "release-date": "2020-01-15",
      "runtime": {"version": "1.1.0"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(3, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-05",
      "runtime": {"version": "1.0.2"}
    },
    {
      "release-date": "2020-02-01",
      "runtime": {"version": "1.0.1"}
    },
    {
      "release-date": "2020-01-10",
      "runtime": {"version": "1.0.0"}
    }
  ]
}
`), nil)

			versions, err := dotnetRuntime.GetAllVersionRefs()
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
      "runtime": {"version": "2.0.1"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-15",
      "runtime": {"version": "1.1.2"}
    },
    {
      "release-date": "2020-02-10",
      "runtime": {"version": "2.0.1"}
    }
  ]
}
`), nil)

			versions, err := dotnetRuntime.GetAllVersionRefs()
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

		it("returns the correct dotnet runtime version", func() {
			fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "runtime": {
        "version": "2.0.2",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.2",
            "hash": "sha512-for-linux-x64-2.0.2"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-20",
      "runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.1",
            "hash": "SHA512-FOR-LINUX-X64-2.0.1"
          },
          {
            "name": "dotnet-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.1",
            "hash": "sha512-for-osx-64-2.0.1"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-10",
      "runtime": {
        "version": "2.0.0",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
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
			fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

			actualDep, err := dotnetRuntime.GetDependencyVersion("2.0.1")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedDep := dependency.DepVersion{
				Version:         "2.0.1",
				URI:             "url-for-linux-x64-2.0.1",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: &expectedDeprecationDate,
				CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.1:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
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

		when("the version is >= 5.0.0", func() {
			it("returns the correct CPE", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "runtime": {
        "version": "5.0.2",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-5.0.2",
            "hash": "sha512-for-linux-x64-5.0.2"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-20",
      "runtime": {
        "version": "5.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-5.0.1",
            "hash": "sha512-for-linux-arm-5.0.1"
          },
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-5.0.1",
            "hash": "SHA512-FOR-LINUX-X64-5.0.1"
          },
          {
            "name": "dotnet-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-5.0.1",
            "hash": "sha512-for-osx-64-5.0.1"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-10",
      "runtime": {
        "version": "5.0.0",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-5.0.0",
            "hash": "sha512-for-linux-x64-5.0.0"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetRuntime.GetDependencyVersion("5.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "5.0.1",
					URI:             "url-for-linux-x64-5.0.1",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net:5.0.1:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/5.0/releases.json", urlArg)

				urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("url-for-linux-x64-5.0.1", urlArg)

				_, sha512Arg := fakeChecksummer.VerifySHA512ArgsForCall(0)
				assert.Equal("sha512-for-linux-x64-5.0.1", sha512Arg)
			})
		})

		when("the file rid is ubuntu-x64", func() {
			it("returns the correct dotnet runtime version", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "dotnet-runtime-ubuntu-x64.tar.gz",
            "rid": "ubuntu-x64",
            "url": "url-for-ubuntu-x64-2.0.1",
            "hash": "SHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA512"
          },
          {
            "name": "dotnet-runtime-osx-64.tar.gz",
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
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetRuntime.GetDependencyVersion("2.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.1",
					URI:             "url-for-ubuntu-x64-2.0.1",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.1:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
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
      "runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "dotnet-runtime-osx-64.tar.gz",
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

				_, err := dotnetRuntime.GetDependencyVersion("2.0.1")
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
      "runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
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
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetRuntime.GetDependencyVersion("2.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.1",
					URI:             "url-for-linux-x64-2.0.1",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.1:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
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
      "runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
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
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetRuntime.GetDependencyVersion("2.0.1")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.1",
					URI:             "url-for-linux-x64-2.0.1",
					SHA256:          "shaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.1:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
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
      "runtime": {
        "version": "2.0.2",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
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
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1")

				actualDep, err := dotnetRuntime.GetDependencyVersion("2.0.2")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.2",
					URI:             "url-for-linux-x64-2.0.2",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.2:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-runtime@2.0.1?checksum=some-sha256&download_url=url-for-linux-x64-2.0.1",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct dotnet runtime release date", func() {
			fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "runtime": {
        "version": "2.0.2",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.2",
            "hash": "sha512-for-linux-x64-2.0.2"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-20",
      "runtime": {
        "version": "2.0.1",
        "files": [
          {
            "name": "dotnet-runtime-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.1",
            "hash": "sha512-for-linux-arm-2.0.1"
          },
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.1",
            "hash": "SHA512-FOR-LINUX-X64-2.0.1"
          },
          {
            "name": "dotnet-runtime-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.1",
            "hash": "sha512-for-osx-64-2.0.1"
          }
        ]
      }
    },
    {
      "release-date": "2020-02-10",
      "runtime": {
        "version": "2.0.0",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
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

			releaseDate, err := dotnetRuntime.GetReleaseDate("2.0.1")
			require.NoError(err)

			assert.Equal("2020-02-20T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
