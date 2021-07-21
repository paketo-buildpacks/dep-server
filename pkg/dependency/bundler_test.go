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

func TestBundler(t *testing.T) {
	spec.Run(t, "Bundler", testBundler, spec.Report(report.Terminal{}))
}

func testBundler(t *testing.T, when spec.G, it spec.S) {
	var (
		assert               = assert.New(t)
		require              = require.New(t)
		fakeChecksummer      *dependencyfakes.FakeChecksummer
		fakeFileSystem       *dependencyfakes.FakeFileSystem
		fakeWebClient        *dependencyfakes.FakeWebClient
		fakeLicenseRetriever *dependencyfakes.FakeLicenseRetriever
		fakePURLGenerator    *dependencyfakes.FakePURLGenerator
		bundler              dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}
		fakeLicenseRetriever = &dependencyfakes.FakeLicenseRetriever{}
		fakePURLGenerator = &dependencyfakes.FakePURLGenerator{}

		var err error
		bundler, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient, fakeLicenseRetriever, fakePURLGenerator).NewDependency("bundler")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all bundler release versions with newest versions first", func() {
			fakeWebClient.GetReturns([]byte(`
[
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-07-02T00:00:00.000Z",
      "created_at":"2020-07-02T12:07:54.097Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":13684,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/rubygems/rubygems/blob/master/bundler/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/rubygems/rubygems/issues?q=is%3Aopen+is%3Aissue+label%3ABundler",
         "source_code_uri":"https://github.com/rubygems/rubygems/"
      },
      "number":"2.2.0.rc.1",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":true,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"2c50355965b6603035ae43199484e9f0f12f45b6d7f7a8d18a503b6d178b4f42"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-01-05T00:00:00.000Z",
      "created_at":"2020-01-05T18:19:06.369Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":44312492,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.4",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"50014d21d6712079da4d6464de12bb93c278f87c9200d0b60ba99f32c25af489"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-01-02T00:00:00.000Z",
      "created_at":"2020-01-02T12:29:43.745Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":438041,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.3",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"9b9a9a5685121403eda1ae148ed3a34c86418f2a2beec7df82a45d4baca0e5d2"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2019-12-20T00:00:00.000Z",
      "created_at":"2019-12-20T00:43:22.951Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":3339529,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.2",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"a3d89c9a7fbfe9364512cac10bc8dc4f9c370e41375c03cd36cad31eef6fb961"
   }
]
`), nil)

			versions, err := bundler.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{"2.1.4", "2.1.3", "2.1.2"}, versions)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct bundler version", func() {
			fakeWebClient.GetReturns([]byte(`
[
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-07-02T00:00:00.000Z",
      "created_at":"2020-07-02T12:07:54.097Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":13684,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/rubygems/rubygems/blob/master/bundler/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/rubygems/rubygems/issues?q=is%3Aopen+is%3Aissue+label%3ABundler",
         "source_code_uri":"https://github.com/rubygems/rubygems/"
      },
      "number":"2.2.0.rc.1",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":true,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"2c50355965b6603035ae43199484e9f0f12f45b6d7f7a8d18a503b6d178b4f42"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-01-05T00:00:00.000Z",
      "created_at":"2020-01-05T18:19:06.369Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":44312492,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.4",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"50014d21d6712079da4d6464de12bb93c278f87c9200d0b60ba99f32c25af489"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-01-02T00:00:00.000Z",
      "created_at":"2020-01-02T12:29:43.745Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":438041,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.3",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"9b9a9a5685121403eda1ae148ed3a34c86418f2a2beec7df82a45d4baca0e5d2"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2019-12-20T00:00:00.000Z",
      "created_at":"2019-12-20T00:43:22.951Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":3339529,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.2",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"a3d89c9a7fbfe9364512cac10bc8dc4f9c370e41375c03cd36cad31eef6fb961"
   }
]
`), nil)

			fakeLicenseRetriever.LookupLicensesReturns([]string{"MIT", "MIT-2"}, nil)
			fakePURLGenerator.GenerateReturns("pkg:generic/bundler@2.1.3?checksum=9b9a9a&download_url=https://rubygems.org")

			actualDepVersion, err := bundler.GetDependencyVersion("2.1.3")
			require.NoError(err)

			assert.Equal(1, fakeLicenseRetriever.LookupLicensesCallCount())
			assert.Equal(1, fakePURLGenerator.GenerateCallCount())
			expectedReleaseDate := time.Date(2020, 01, 02, 12, 29, 43, 745000000, time.UTC)
			expectedDepVersion := dependency.DepVersion{
				Version:         "2.1.3",
				URI:             "https://rubygems.org/downloads/bundler-2.1.3.gem",
				SHA256:          "9b9a9a5685121403eda1ae148ed3a34c86418f2a2beec7df82a45d4baca0e5d2",
				ReleaseDate:     &expectedReleaseDate,
				DeprecationDate: nil,
				CPE:             "cpe:2.3:a:bundler:bundler:2.1.3:*:*:*:*:ruby:*:*",
				PURL:            "pkg:generic/bundler@2.1.3?checksum=9b9a9a&download_url=https://rubygems.org",
				Licenses:        []string{"MIT", "MIT-2"},
			}
			assert.Equal(expectedDepVersion, actualDepVersion)

			url, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://rubygems.org/api/v1/versions/bundler.json", url)
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct bundler release date", func() {
			fakeWebClient.GetReturns([]byte(`
[
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-07-02T00:00:00.000Z",
      "created_at":"2020-07-02T12:07:54.097Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":13684,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/rubygems/rubygems/blob/master/bundler/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/rubygems/rubygems/issues?q=is%3Aopen+is%3Aissue+label%3ABundler",
         "source_code_uri":"https://github.com/rubygems/rubygems/"
      },
      "number":"2.2.0.rc.1",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":true,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"2c50355965b6603035ae43199484e9f0f12f45b6d7f7a8d18a503b6d178b4f42"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-01-05T00:00:00.000Z",
      "created_at":"2020-01-05T18:19:06.369Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":44312492,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.4",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"50014d21d6712079da4d6464de12bb93c278f87c9200d0b60ba99f32c25af489"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2020-01-02T00:00:00.000Z",
      "created_at":"2020-01-02T12:29:43.745Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":438041,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.3",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"9b9a9a5685121403eda1ae148ed3a34c86418f2a2beec7df82a45d4baca0e5d2"
   },
   {
      "authors":"André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
      "built_at":"2019-12-20T00:00:00.000Z",
      "created_at":"2019-12-20T00:43:22.951Z",
      "description":"Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
      "downloads_count":3339529,
      "metadata":{
         "homepage_uri":"https://bundler.io/",
         "changelog_uri":"https://github.com/bundler/bundler/blob/master/CHANGELOG.md",
         "bug_tracker_uri":"https://github.com/bundler/bundler/issues",
         "source_code_uri":"https://github.com/bundler/bundler/"
      },
      "number":"2.1.2",
      "summary":"The best way to manage your application's dependencies",
      "platform":"ruby",
      "rubygems_version":"\u003e= 2.5.2",
      "ruby_version":"\u003e= 2.3.0",
      "prerelease":false,
      "licenses":[
         "MIT"
      ],
      "requirements":[

      ],
      "sha":"a3d89c9a7fbfe9364512cac10bc8dc4f9c370e41375c03cd36cad31eef6fb961"
   }
]
`), nil)

			releaseDate, err := bundler.GetReleaseDate("2.1.3")
			require.NoError(err)

			assert.Equal("2020-01-02T12:29:43.745Z", releaseDate.Format(time.RFC3339Nano))
		})
	})
}
