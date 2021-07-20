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

func TestCurl(t *testing.T) {
	spec.Run(t, "Curl", testCurl, spec.Report(report.Terminal{}))
}

func testCurl(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		curl                 dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		curl, err = dependency.NewCustomDependencyFactory(fakeChecksummer, nil, nil, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("curl")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all curl release versions ordered with newest versions first", func() {
			fakeWebClient.GetReturns([]byte(`
0;7.74.0;0;2020-12-09;0;56;56;107;107;1;1;
1;7.73.0;3;2020-10-14;56 days;56;112;135;242;9;10;
2;7.72.0;3;2020-08-19;3 months;49;161;100;342;3;13;
3;7.71.1;4;2020-07-01;5 months;7;168;18;360;0;13;
4;7.71.0;4;2020-06-24;5 months;56;224;136;496;4;17;
`), nil)

			versions, err := curl.GetAllVersionRefs()
			require.NoError(err)

			assert.Equal([]string{
				"7.74.0",
				"7.73.0",
				"7.72.0",
				"7.71.1",
				"7.71.0",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://curl.se/docs/releases.csv", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct curl version", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
0;7.74.0;0;2020-12-09;0;56;56;107;107;1;1;
1;7.73.0;3;2020-10-14;56 days;56;112;135;242;9;10;
2;7.72.0;3;2020-08-19;3 months;49;161;100;342;3;13;
`), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte("some-gpg-key"), nil)
			fakeWebClient.GetReturnsOnCall(2, []byte("some-signature"), nil)
			fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/curl@7.73.0?checksum=some-source-sha&download_url=https://curl.se")

			actualDep, err := curl.GetDependencyVersion("7.73.0")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())

			expectedReleaseDate := time.Date(2020, 10, 14, 0, 0, 0, 0, time.UTC)
			expectedDep := dependency.DepVersion{
				Version:     "7.73.0",
				URI:         "https://curl.se/download/curl-7.73.0.tar.gz",
				SHA256:      "some-source-sha",
				ReleaseDate: &expectedReleaseDate,
				CPE:         "cpe:2.3:a:haxx:curl:7.73.0:*:*:*:*:*:*:*",
				PURL:        "pkg:generic/curl@7.73.0?checksum=some-source-sha&download_url=https://curl.se",
				Licenses:    []string{"MIT", "MIT-2"},
			}

			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://curl.se/docs/releases.csv", urlArg)

			urlArg, _ = fakeWebClient.GetArgsForCall(1)
			assert.Equal("https://daniel.haxx.se/mykey.asc", urlArg)

			urlArg, _ = fakeWebClient.GetArgsForCall(2)
			assert.Equal("https://curl.se/download/curl-7.73.0.tar.gz.asc", urlArg)

			urlArg, _, _ = fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("https://curl.se/download/curl-7.73.0.tar.gz", urlArg)

			releaseAssetSignatureArg, _, curlGPGKeyArg := fakeChecksummer.VerifyASCArgsForCall(0)
			assert.Equal("some-signature", releaseAssetSignatureArg)
			assert.Equal([]string{"some-gpg-key"}, curlGPGKeyArg)
		})

		when("the version is older then 7.3.30", func() {
			it("uses the 'archeology' download URL and does not verify the signature", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
62;7.31.0;55;2013-06-22;7.5 years;71;2798;39;4406;8;282;
63;7.30.0;53;2013-04-12;7.7 years;65;2863;50;4456;13;295;
64;7.29.0;52;2013-02-06;7.8 years;78;2941;35;4491;10;305;
65;7.28.1;52;2012-11-20;8.1 years;41;2982;31;4522;3;308;
`), nil)
				fakeChecksummer.GetSHA256Returns("some-source-sha", nil)
				fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
				fakePURLGenerator.GenerateReturns("pkg:generic/curl@7.29.0?checksum=some-source-sha&download_url=https://curl.se")

				actualDep, err := curl.GetDependencyVersion("7.29.0")
				require.NoError(err)

				assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
				assert.Equal(1, fakePURLGenerator.GenerateCallCount())

				expectedReleaseDate := time.Date(2013, 02, 06, 0, 0, 0, 0, time.UTC)
				expectedDep := dependency.DepVersion{
					Version:     "7.29.0",
					URI:         "https://curl.se/download/archeology/curl-7.29.0.tar.gz",
					SHA256:      "some-source-sha",
					ReleaseDate: &expectedReleaseDate,
					CPE:         "cpe:2.3:a:haxx:curl:7.29.0:*:*:*:*:*:*:*",
					PURL:        "pkg:generic/curl@7.29.0?checksum=some-source-sha&download_url=https://curl.se",
					Licenses:    []string{"MIT", "MIT-2"},
				}

				assert.Equal(expectedDep, actualDep)

				assert.Equal(0, fakeChecksummer.VerifyASCCallCount())
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct curl release date", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
0;7.74.0;0;2020-12-09;0;56;56;107;107;1;1;
1;7.73.0;3;2020-10-14;56 days;56;112;135;242;9;10;
2;7.72.0;3;2020-08-19;3 months;49;161;100;342;3;13;
`), nil)

			releaseDate, err := curl.GetReleaseDate("7.73.0")
			require.NoError(err)

			assert.Equal("2020-10-14T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
