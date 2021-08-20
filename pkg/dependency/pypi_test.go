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

func TestPyPi(t *testing.T) {
	spec.Run(t, "PyPi", testPyPi, spec.Report(report.Terminal{}))
}

func testPyPi(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeGithubClient     *dependencyfakes.FakeGithubClient
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		pypi                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeGithubClient = &dependencyfakes.FakeGithubClient{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		pypi, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("pip")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all source release final versions with newest versions first", func() {
			fakeWebClient.GetReturns([]byte(`{
"releases": {
  "1.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-01-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "1.0.1": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-03-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "1.10.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "1.2.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-03-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "1.20.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-08-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.0.0.dev2": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.0.0b1": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.0.0": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.1.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-08-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}},
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-08-01T00:00:00.000000Z"}
  ]
}}`), nil)
			versions, err := pypi.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{
				"2.1.0",
				"1.20.0",
				"2.0.0",
				"1.10.0",
				"1.2.0",
				"1.0.1",
				"1.0.0",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://pypi.org/pypi/pip/json", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct version", func() {
			fakeWebClient.GetReturns([]byte(`{
"releases": {
  "1.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-01-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.0.0": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {
      "packagetype": "sdist",
      "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z",
      "digests": {
        "md5": "some-md5",
        "sha256": "some-sha256"
      },
      "url": "some-url"
    }
  ],
  "3.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-08-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ]
}}`), nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/pip@2.0.0?checksum=some-sha-256gz&download_url=some-url")

			actualDep, err := pypi.GetDependencyVersion("2.0.0")
			require.NoError(err)

			assert.Equal(3, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(3, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2010, 5, 1, 0, 0, 0, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:         "2.0.0",
				URI:             "some-url",
				SHA256:          "some-sha256",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:pypa:pip:2.0.0:*:*:*:*:python:*:*",
				PURL:            "pkg:generic/pip@2.0.0?checksum=some-sha-256gz&download_url=some-url",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://pypi.org/pypi/pip/json", urlArg)
		})

		when("the product is not pip", func() {
			it.Before(func() {
				var err error
				pypi, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, fakeGithubClient, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("pipenv")
				require.NoError(err)
			})
			it("the CPE field is empty", func() {
				fakeWebClient.GetReturns([]byte(`{
"releases": {
  "1.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-01-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.0.0": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {
      "packagetype": "sdist",
      "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z",
      "digests": {
        "md5": "some-md5",
        "sha256": "some-sha256"
      },
      "url": "some-url"
    }
  ],
  "3.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-08-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ]
}}`), nil)

				actualDep, err := pypi.GetDependencyVersion("2.0.0")
				require.NoError(err)

				assert.Equal("", actualDep.CPE)
			})
		})

		when("the sha256 is empty", func() {
			it("returns an error", func() {
				fakeWebClient.GetReturns([]byte(`{
"releases": {
  "2.0.0": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {
      "packagetype": "sdist",
      "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z",
      "digests": {
        "md5": "some-md5"
      },
      "url": "some-url"
    }
  ]
}}`), nil)

				_, err := pypi.GetDependencyVersion("2.0.0")
				assert.Error(err)
				assert.Contains(err.Error(), "could not find sha256 for version 2.0.0")
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct release date", func() {
			fakeWebClient.GetReturns([]byte(`{
"releases": {
  "1.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-01-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ],
  "2.0.0": [
    {"packagetype": "bdist_wheel", "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z"},
    {
      "packagetype": "sdist",
      "upload_time_iso_8601": "2010-05-01T00:00:00.000000Z",
      "digests": {
        "md5": "some-md5",
        "sha256": "some-sha256"
      },
      "url": "some-url"
    }
  ],
  "3.0.0": [
    {"packagetype": "sdist", "upload_time_iso_8601": "2010-08-01T00:00:00.000000Z", "digests": {"sha256": "some-sha256"}}
  ]
}}`), nil)

			releaseDate, err := pypi.GetReleaseDate("2.0.0")
			require.NoError(err)

			assert.Equal("2010-05-01T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
