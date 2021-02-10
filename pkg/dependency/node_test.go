package dependency_test

import (
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/dependencyfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNode(t *testing.T) {
	spec.Run(t, "Node", testNode, spec.Report(report.Terminal{}))
}

func testNode(t *testing.T, when spec.G, it spec.S) {
	var (
		assert          = assert.New(t)
		require         = require.New(t)
		fakeChecksummer *dependencyfakes.FakeChecksummer
		fakeFileSystem  *dependencyfakes.FakeFileSystem
		fakeWebClient   *dependencyfakes.FakeWebClient
		node            dependency.Dependency
	)

	it.Before(func() {
		fakeChecksummer = &dependencyfakes.FakeChecksummer{}
		fakeFileSystem = &dependencyfakes.FakeFileSystem{}
		fakeWebClient = &dependencyfakes.FakeWebClient{}

		var err error
		node, err = dependency.NewCustomDependencyFactory(fakeChecksummer, fakeFileSystem, nil, fakeWebClient).NewDependency("node")
		require.NoError(err)
	})

	when("GetAllVersionRefs", func() {
		it("returns all node release versions ordered with newest versions first, then highest semver first", func() {
			fakeWebClient.GetReturns([]byte(`
[
  {"version": "v14.0.0", "date": "2020-01-30"},
  {"version": "v13.9.0", "date": "2020-02-18"},
  {"version": "v13.8.0", "date": "2020-02-02"},
  {"version": "v13.7.0", "date": "2020-01-30"},
  {"version": "v13.6.1", "date": "2020-01-30"},
  {"version": "v12.16.1", "date": "2020-02-15"},
  {"version": "v12.16.0", "date": "2020-02-10"},
  {"version": "v12.15.0", "date": "2020-01-30"},
  {"version": "v10.19.0", "date": "2020-01-30"}
]`), nil)

			versions, err := node.GetAllVersionRefs()

			require.NoError(err)

			assert.Equal([]string{
				"v13.9.0",
				"v12.16.1",
				"v12.16.0",
				"v13.8.0",
				"v14.0.0",
				"v13.7.0",
				"v13.6.1",
				"v12.15.0",
				"v10.19.0",
			}, versions)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://nodejs.org/dist/index.json", urlArg)
		})
	})

	when("GetDependencyVersion", func() {
		it("returns the correct node version", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
[
 {"version": "v14.0.0", "date": "2020-01-30"},
 {"version": "v13.9.0", "date": "2020-02-20"},
 {"version": "v13.8.0", "date": "2020-02-20"}
]`), nil)

			fakeWebClient.GetReturnsOnCall(1, []byte(`
{
 "v13": {
   "start": "2019-10-22",
   "maintenance": "2020-04-01",
   "end": "2020-06-01"
 },
 "v14": {
   "start": "2020-04-21",
   "lts": "2020-10-20",
   "maintenance": "2021-10-19",
   "end": "2023-04-30",
   "codename": ""
 }
}`), nil)

			fakeWebClient.GetReturnsOnCall(2, []byte(`
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  node-v13.9.0.tar.gz
`), nil)

			actualDep, err := node.GetDependencyVersion("v13.9.0")

			require.NoError(err)

			expectedDep := dependency.DepVersion{
				Version:         "v13.9.0",
				URI:             "https://nodejs.org/dist/v13.9.0/node-v13.9.0.tar.gz",
				SHA:             "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ReleaseDate:     "2020-02-20T00:00:00Z",
				DeprecationDate: "2020-06-01T00:00:00Z",
			}

			assert.Equal(expectedDep, actualDep)

			urlArg, _ := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://nodejs.org/dist/index.json", urlArg)
		})

		when("the major version is 0", func() {
			it("pulls the correct release schedule", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
		[
		 {"version": "v0.8.0", "date": "2015-01-30"}
		]`), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(`
{
 "v0.8": {
   "start": "2015-01-30",
   "maintenance": "2015-04-01",
   "end": "2016-06-01"
 }
}`), nil)

				fakeWebClient.GetReturnsOnCall(2, []byte(`
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  node-v0.8.0.tar.gz
`), nil)

				actualDep, err := node.GetDependencyVersion("v0.8.0")

				require.NoError(err)

				expectedDep := dependency.DepVersion{
					Version:         "v0.8.0",
					URI:             "https://nodejs.org/dist/v0.8.0/node-v0.8.0.tar.gz",
					SHA:             "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					ReleaseDate:     "2015-01-30T00:00:00Z",
					DeprecationDate: "2016-06-01T00:00:00Z",
				}

				assert.Equal(expectedDep, actualDep)
			})
		})

		when("the deprecation date cannot be found", func() {
			it("returns the dependency with an empty deprecation date", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
		[
		 {"version": "v13.9.0", "date": "2020-01-30"}
		]`), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(`
{
 "v0.8": {
   "start": "2015-01-30",
   "maintenance": "2015-04-01",
   "end": "2016-06-01"
 }
}`), nil)

				fakeWebClient.GetReturnsOnCall(2, []byte(`
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  node-v13.9.0.tar.gz
`), nil)

				actualDep, err := node.GetDependencyVersion("v13.9.0")

				require.NoError(err)

				expectedDep := dependency.DepVersion{
					Version:         "v13.9.0",
					URI:             "https://nodejs.org/dist/v13.9.0/node-v13.9.0.tar.gz",
					SHA:             "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					ReleaseDate:     "2020-01-30T00:00:00Z",
					DeprecationDate: "",
				}

				assert.Equal(expectedDep, actualDep)
			})
		})

		when("the release schedule cannot be found", func() {
			it("returns an error", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
		[
		 {"version": "v14.0.0", "date": "2020-01-30"}
		]`), nil)
				fakeWebClient.GetReturnsOnCall(1, nil, nil)

				_, err := node.GetDependencyVersion("v14.0.0")
				assert.Error(err)
				assert.Contains(err.Error(), "could not unmarshal release schedule:")
			})
		})

		when("the SHA cannot be found", func() {
			it("returns an error", func() {
				fakeWebClient.GetReturnsOnCall(0, []byte(`
		[
		 {"version": "v14.0.0", "date": "2020-01-30"}
		]`), nil)
				fakeWebClient.GetReturnsOnCall(1, []byte(`
		{
		 "v14": {
		   "start": "2020-04-21",
		   "lts": "2020-10-20",
		   "maintenance": "2021-10-19",
		   "end": "2023-04-30",
		   "codename": ""
		 }
		}`), nil)
				fakeWebClient.GetReturnsOnCall(2, nil, nil)
				_, err := node.GetDependencyVersion("v14.0.0")
				assert.Error(err)
				assert.Contains(err.Error(), "could not find SHA for node-v14.0.0.tar.gz")
			})
		})
	})

	when("GetReleaseDate", func() {
		it("returns the correct node release date", func() {
			fakeWebClient.GetReturnsOnCall(0, []byte(`
[
 {"version": "v14.0.0", "date": "2020-01-30"},
 {"version": "v13.9.0", "date": "2020-02-20"},
 {"version": "v13.8.0", "date": "2020-02-20"}
]`), nil)

			releaseDate, err := node.GetReleaseDate("v13.9.0")
			require.NoError(err)

			assert.Equal("2020-02-20T00:00:00Z", releaseDate.Format(time.RFC3339))
		})
	})
}
