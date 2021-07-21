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

func TestDotnetSDK(t *testing.T) {
	spec.Run(t, "Dotnet SDK", testDotnetSDK, spec.Report(report.Terminal{}))
}

func testDotnetSDK(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		dotnetSDK            dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		dotnetSDK, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("dotnet-sdk")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all dotnet SDK final versions with newest versions first", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{"releases-index": [{"channel-version": "2.0"}, {"channel-version": "1.1"}, {"channel-version": "1.0"}]}
`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "sdk": {"version": "2.0.201"},
      "sdks": [{"version": "2.0.201"}, {"version": "2.0.101"}]
    },
    {
      "release-date": "2020-02-10",
      "sdk": {"version": "2.0.200"},
      "sdks": [{"version": "2.0.200"}]
    },
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "2.0.100"}
    },
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "2.0.100-preview1.20000.20"}
    },
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "2.0.100-preview1-20000-20"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-10",
      "sdk": {"version": "1.1.201"},
      "sdks": [{"version": "1.1.201"}, {"version": "1.1.101"}]
    },
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "1.1.200"},
      "sdks": [{"version": "1.1.200"}]
    },
    {
      "release-date": "2020-01-15",
      "sdk": {"version": "1.1.100"}
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(3, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-05",
      "sdk": {"version": "1.0.201"},
      "sdks": [{"version": "1.0.201"}, {"version": "1.0.101"}]
    },
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "1.0.200"},
      "sdks": [{"version": "1.0.200"}]
    },
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "1.0.200"}
    },
    {
      "release-date": "2020-01-10",
      "sdk": {"version": "1.0.100"}
    }
  ]
}
`), nil)

			versions, err := dotnetSDK.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"2.0.201",
				"2.0.101",
				"2.0.200",
				"1.1.201",
				"1.1.101",
				"1.0.201",
				"1.0.101",
				"2.0.100",
				"1.1.200",
				"1.0.200",
				"1.1.100",
				"1.0.100",
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
      "sdk": {"version": "2.0.100"},
      "sdks": [{"version": "2.0.100"}]
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-15",
      "sdk": {"version": "1.1.200"},
      "sdks": [{"version": "1.1.200"}]
    },
    {
      "release-date": "2020-02-10",
      "sdk": {"version": "2.0.100"},
      "sdks": [{"version": "2.0.100"}]
    }
  ]
}
`), nil)

			versions, err := dotnetSDK.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"1.1.200",
				"2.0.100",
			}, versions)
		})

		it("excludes version 2.1.202 which has the wrong hash", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
{"releases-index": [{"channel-version": "2.1"}, {"channel-version": "2.0"}]}
`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-01",
      "sdk": {"version": "2.1.100"},
      "sdks": [{"version": "2.1.100"}]
    }
  ]
}
`), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte(`
{
  "releases": [
    {
      "release-date": "2020-02-20",
      "sdk": {"version": "2.1.202"},
      "sdks": [{"version": "2.1.202"}]
    },
    {
      "release-date": "2020-01-15",
      "sdk": {"version": "2.0.100"},
      "sdks": [{"version": "2.0.100"}]
    }
  ]
}
`), nil)

			versions, err := dotnetSDK.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"2.1.100",
				"2.0.100",
			}, versions)
		})
	})

	when("GetDependencyVersion", func() {
		var (
			expectedReleaseDate     = time.Date(2020, 02, 20, 0, 0, 0, 0, time.UTC)
			expectedDeprecationDate = time.Date(2050, 02, 20, 0, 0, 0, 0, time.UTC)
		)

		it("returns the correct dotnet SDK version", func() {
			fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "sdk": {
        "version": "2.0.202",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.202",
            "hash": "sha512-for-linux-x64-2.0.202"
          }
        ]
      },
      "sdks": [
        {
          "version": "2.0.202",
          "files": [
            {
              "name": "dotnet-sdk-linux-x64.tar.gz",
              "rid": "linux-x64",
              "url": "url-for-linux-x64-2.0.202",
              "hash": "sha512-for-linux-x64-2.0.202"
            }
          ]
        }
      ]
    },
    {
      "release-date": "2020-02-20",
      "sdk": {
        "version": "2.0.301"
      },
      "sdks": [
        {
          "version": "2.0.301"
        },
        {
          "version": "2.0.201",
          "files": [
            {
              "name": "dotnet-sdk-linux-arm.tar.gz",
              "rid": "linux-arm",
              "url": "url-for-linux-arm-2.0.201",
              "hash": "sha512-for-linux-arm-2.0.201"
            },
            {
              "name": "dotnet-sdk-linux-x64.tar.gz",
              "rid": "linux-x64",
              "url": "url-for-linux-x64-2.0.201",
              "hash": "SHA512-FOR-LINUX-X64-2.0.201"
            },
            {
              "name": "dotnet-sdk-osx-64.tar.gz",
              "rid": "osx-64",
              "url": "url-for-osx-64-2.0.201",
              "hash": "sha512-for-osx-64-2.0.201"
            }
          ]
        }
      ]
    },
    {
      "release-date": "2020-02-10",
      "sdk": {
        "version": "2.0.200",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.200",
            "hash": "sha512-for-linux-x64-2.0.200"
          }
        ]
      }
    }
  ]
}
`), nil)
			fakeChecksummer.GetSHA256Returns("some-sha256", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201")

			actualDep, err := dotnetSDK.GetDependencyVersion("2.0.201")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedDep := dependency.DepVersion{
				Version:         "2.0.201",
				URI:             "url-for-linux-x64-2.0.201",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: &expectedDeprecationDate,
				CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.201:*:*:*:*:*:*:*",
				PURL:            "pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/2.0/releases.json", urlArg)

			urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("url-for-linux-x64-2.0.201", urlArg)

			_, sha512Arg := fakeChecksummer.VerifySHA512ArgsForCall(0)
			assert.Equal("sha512-for-linux-x64-2.0.201", sha512Arg)
		})

		when("the version is >= 5.0.0", func() {
			it("returns the correct CPE", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "sdk": {
        "version": "5.0.202",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-5.0.202",
            "hash": "sha512-for-linux-x64-5.0.202"
          }
        ]
      },
      "sdks": [
        {
          "version": "5.0.202",
          "files": [
            {
              "name": "dotnet-sdk-linux-x64.tar.gz",
              "rid": "linux-x64",
              "url": "url-for-linux-x64-5.0.202",
              "hash": "sha512-for-linux-x64-5.0.202"
            }
          ]
        }
      ]
    },
    {
      "release-date": "2020-02-20",
      "sdk": {
        "version": "5.0.301"
      },
      "sdks": [
        {
          "version": "5.0.301"
        },
        {
          "version": "5.0.201",
          "files": [
            {
              "name": "dotnet-sdk-linux-arm.tar.gz",
              "rid": "linux-arm",
              "url": "url-for-linux-arm-5.0.201",
              "hash": "sha512-for-linux-arm-5.0.201"
            },
            {
              "name": "dotnet-sdk-linux-x64.tar.gz",
              "rid": "linux-x64",
              "url": "url-for-linux-x64-5.0.201",
              "hash": "SHA512-FOR-LINUX-X64-5.0.201"
            },
            {
              "name": "dotnet-sdk-osx-64.tar.gz",
              "rid": "osx-64",
              "url": "url-for-osx-64-5.0.201",
              "hash": "sha512-for-osx-64-5.0.201"
            }
          ]
        }
      ]
    },
    {
      "release-date": "2020-02-10",
      "sdk": {
        "version": "5.0.200",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-5.0.200",
            "hash": "sha512-for-linux-x64-5.0.200"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@5.0.201?checksum=some-sha256&download_url=url-for-linux-x64-5.0.201")

				actualDep, err := dotnetSDK.GetDependencyVersion("5.0.201")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "5.0.201",
					URI:             "url-for-linux-x64-5.0.201",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net:5.0.201:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-sdk@5.0.201?checksum=some-sha256&download_url=url-for-linux-x64-5.0.201",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/5.0/releases.json", urlArg)

				urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("url-for-linux-x64-5.0.201", urlArg)

				_, sha512Arg := fakeChecksummer.VerifySHA512ArgsForCall(0)
				assert.Equal("sha512-for-linux-x64-5.0.201", sha512Arg)
			})
		})

		when("the file rid is ubuntu-x64", func() {
			it("returns the correct dotnet SDK version", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "sdk": {
        "version": "2.0.201",
        "files": [
          {
            "name": "dotnet-sdk-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.201",
            "hash": "sha512-for-linux-arm-2.0.201"
          },
          {
            "name": "dotnet-sdk-ubuntu-x64.tar.gz",
            "rid": "ubuntu-x64",
            "url": "url-for-ubuntu-x64-2.0.201",
            "hash": "SHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA512"
          },
          {
            "name": "dotnet-sdk-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.201",
            "hash": "sha512-for-osx-64-2.0.201"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201")

				actualDep, err := dotnetSDK.GetDependencyVersion("2.0.201")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.201",
					URI:             "url-for-ubuntu-x64-2.0.201",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.201:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/2.0/releases.json", urlArg)

				urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
				assert.Equal("url-for-ubuntu-x64-2.0.201", urlArg)

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
      "sdk": {
        "version": "2.0.201",
        "files": [
          {
            "name": "dotnet-sdk-linux-arm.tar.gz",
            "rid": "linux-arm",
            "url": "url-for-linux-arm-2.0.201",
            "hash": "sha512-for-linux-arm-2.0.201"
          },
          {
            "name": "dotnet-sdk-osx-64.tar.gz",
            "rid": "osx-64",
            "url": "url-for-osx-64-2.0.201",
            "hash": "sha512-for-osx-64-2.0.201"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)

				_, err := dotnetSDK.GetDependencyVersion("2.0.201")
				assert.Error(err)

				assert.True(errors.Is(err, depErrors.NoSourceCodeError{Version: "2.0.201"}))
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
      "sdk": {
        "version": "2.0.201",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.201",
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
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201")

				actualDep, err := dotnetSDK.GetDependencyVersion("2.0.201")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.201",
					URI:             "url-for-linux-x64-2.0.201",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.201:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201",
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
      "sdk": {
        "version": "2.0.201",
        "files": [
          {
            "name": "dotnet-runtime-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.201",
            "hash": "SHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA256"
          }
        ]
      }
    }
  ]
}
`), nil)

				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201")

				actualDep, err := dotnetSDK.GetDependencyVersion("2.0.201")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.201",
					URI:             "url-for-linux-x64-2.0.201",
					SHA256:          "shaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.201:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				assert.Equal(0, fakeWebClient.DownloadCallCount())
				assert.Equal(0, fakeChecksummer.VerifySHA512CallCount())
			})
		})

		when("the version is in the wrong channel", func() {
			it("uses the correct channel", func() {
				fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-20",
      "sdk": {
        "version": "2.1.201",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.1.201",
            "hash": "sha512-for-linux-x64-2.1.201"
          }
        ]
      }
    }
  ]
}
`), nil)
				fakeChecksummer.GetSHA256Returns("some-sha256", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201")

				actualDep, err := dotnetSDK.GetDependencyVersion("2.1.201")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.1.201",
					URI:             "url-for-linux-x64-2.1.201",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: &expectedDeprecationDate,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.1.201:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)

				urlArg, _ := fakeWebClient.GetArgsForCall(0)
				assert.Equal("https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/2.0/releases.json", urlArg)
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
      "sdk": {
        "version": "2.0.201",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.201",
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
				fakePURLGenerator.GenerateReturns("pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201")

				actualDep, err := dotnetSDK.GetDependencyVersion("2.0.201")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())
				expectedDep := dependency.DepVersion{
					Version:         "2.0.201",
					URI:             "url-for-linux-x64-2.0.201",
					SHA256:          "some-sha256",
					ReleaseDate:     &expectedReleaseDate,
					DeprecationDate: nil,
					CPE:             "cpe:2.3:a:microsoft:.net_core:2.0.201:*:*:*:*:*:*:*",
					PURL:            "pkg:generic/dotnet-sdk@2.0.201?checksum=some-sha256&download_url=url-for-linux-x64-2.0.201",
					Licenses:        []string{"MIT", "MIT-2"},
				}
				assert.Equal(expectedDep, actualDep)
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct dotnet SDK release date", func() {
			fakeWebClient.GetReturns([]byte(`
{
  "eol-date": "2050-02-20",
  "releases": [
    {
      "release-date": "2020-02-22",
      "sdk": {
        "version": "2.0.202",
				"files": [
					{
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.202",
            "hash": "sha512-for-linux-x64-2.0.202"
          }
        ]
      },
      "sdks": [
        {
          "version": "2.0.202",
          "files": [
            {
              "name": "dotnet-sdk-linux-x64.tar.gz",
              "rid": "linux-x64",
              "url": "url-for-linux-x64-2.0.202",
              "hash": "sha512-for-linux-x64-2.0.202"
            }
          ]
        }
      ]
    },
    {
      "release-date": "2020-02-20",
      "sdk": {
        "version": "2.0.301"
      },
      "sdks": [
        {
          "version": "2.0.301"
        },
        {
          "version": "2.0.201",
          "files": [
            {
              "name": "dotnet-sdk-linux-arm.tar.gz",
              "rid": "linux-arm",
              "url": "url-for-linux-arm-2.0.201",
              "hash": "sha512-for-linux-arm-2.0.201"
            },
            {
              "name": "dotnet-sdk-linux-x64.tar.gz",
              "rid": "linux-x64",
              "url": "url-for-linux-x64-2.0.201",
              "hash": "SHA512-FOR-LINUX-X64-2.0.201"
            },
            {
              "name": "dotnet-sdk-osx-64.tar.gz",
              "rid": "osx-64",
              "url": "url-for-osx-64-2.0.201",
              "hash": "sha512-for-osx-64-2.0.201"
            }
          ]
        }
      ]
    },
    {
      "release-date": "2020-02-10",
      "sdk": {
        "version": "2.0.200",
        "files": [
          {
            "name": "dotnet-sdk-linux-x64.tar.gz",
            "rid": "linux-x64",
            "url": "url-for-linux-x64-2.0.200",
            "hash": "sha512-for-linux-x64-2.0.200"
          }
        ]
      }
    }
  ]
}
`), nil)
			fakeChecksummer.GetSHA256Returns("some-sha256", nil)

			releaseDate, err := dotnetSDK.GetReleaseDate("2.0.201")
			require.NoError(err)

			assert.Equal("2020-02-20T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
