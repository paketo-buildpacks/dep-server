package internal_test

import (
	"errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internalfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestGithubClient(t *testing.T) {
	spec.Run(t, "GithubClient", testGithubClient, spec.Report(report.Terminal{}))
}

func testGithubClient(t *testing.T, when spec.G, it spec.S) {
	var (
		assert        = assert.New(t)
		require       = require.New(t)
		fakeWebClient *internalfakes.FakeGithubWebClient
		githubClient  internal.GithubClient
	)

	it.Before(func() {
		fakeWebClient = &internalfakes.FakeGithubWebClient{}
		githubClient = internal.NewGithubClient(fakeWebClient, "some-access-token")
	})

	when("GetReleaseTags", func() {
		it("lists tags for all non-draft and non-prerelease releases", func() {
			releasesResponse := `[
	{"tag_name": "v0.0.0", "draft": true, "published_at": "2020-06-30T00:00:00Z"},
	{"tag_name": "v0.0.1", "prerelease": true, "published_at": "2020-06-29T00:00:00Z"},
	{"tag_name": "v1.0.1", "published_at": "2020-06-28T00:00:00Z"},
	{"tag_name": "v2.0.0", "published_at": "2020-06-27T00:00:00Z"},
	{"tag_name": "v1.0.0", "published_at": "2020-06-26T00:00:00Z"}
]`
			fakeWebClient.GetReturns([]byte(releasesResponse), nil)

			actualReleases, err := githubClient.GetReleaseTags("some-org", "some-repo")
			require.NoError(err)

			expectedRelases := []internal.GithubRelease{
				{
					TagName:       "v1.0.1",
					PublishedDate: "2020-06-28T00:00:00Z",
				},
				{
					TagName:       "v2.0.0",
					PublishedDate: "2020-06-27T00:00:00Z",
				},
				{
					TagName:       "v1.0.0",
					PublishedDate: "2020-06-26T00:00:00Z",
				},
			}
			assert.Equal(expectedRelases, actualReleases)

			urlArg, optionsArg := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://api.github.com/repos/some-org/some-repo/releases", urlArg)
			assert.Len(optionsArg, 1)

			request, err := http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))
		})
	})

	when("GetTags", func() {
		it("lists all tags using the GraphQL API", func() {
			tagsResponse := `
{
  "data": {
    "repository": {
      "refs": {
        "edges": [
          {"node": {"name": "1.0.1"}},
          {"node": {"name": "2.0.0"}},
          {"node": {"name": "1.0.0"}}
        ]
      }
    }
  }
}
`
			fakeWebClient.PostReturns([]byte(tagsResponse), nil)

			tags, err := githubClient.GetTags("some-org", "some-repo")
			require.NoError(err)
			assert.Equal([]string{"1.0.1", "2.0.0", "1.0.0"}, tags)

			urlArg, bodyArg, optionsArg := fakeWebClient.PostArgsForCall(0)
			assert.Equal("https://api.github.com/graphql", urlArg)
			assert.Contains(string(bodyArg), `repository(owner: \"some-org\", name: \"some-repo\")`)
			assert.Len(optionsArg, 1)

			request, err := http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))
		})
	})

	when("GetReleaseAsset", func() {
		it("returns the contents of a file for a given release tag", func() {
			releasesResponse := `
{
	"tag_name": "some-tag",
	"assets": [
		{"name": "some-other-asset-name", "url": "some-other-asset-url"},
		{"name": "some-asset-name", "url": "some-asset-url"}
	]
}
`
			fakeWebClient.GetReturnsOnCall(0, []byte(releasesResponse), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte("some-contents"), nil)

			contents, err := githubClient.GetReleaseAsset("some-org", "some-repo", "some-tag", "some-asset-name")
			require.NoError(err)
			assert.Equal([]byte("some-contents"), contents)

			urlArg, optionsArg := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://api.github.com/repos/some-org/some-repo/releases/tags/some-tag", urlArg)
			assert.Len(optionsArg, 1)

			request, err := http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))

			urlArg, optionsArg = fakeWebClient.GetArgsForCall(1)
			assert.Equal("some-asset-url", urlArg)
			assert.Len(optionsArg, 2)

			request, err = http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			optionsArg[1](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))
			assert.Equal("application/octet-stream", request.Header.Get("Accept"))
		})

		when("the asset name does not exist in the release", func() {
			it("returns an AssetNotFound error", func() {
				releasesResponse := `{"tag_name": "some-tag", "assets": []}`
				fakeWebClient.GetReturnsOnCall(0, []byte(releasesResponse), nil)

				_, err := githubClient.GetReleaseAsset("some-org", "some-repo", "some-tag", "some-asset-name")
				assert.Error(err)
				assert.True(errors.Is(err, internal_errors.AssetNotFound{AssetName: "some-asset-name"}))
			})
		})
	})

	when("DownloadReleaseAsset", func() {
		it("downloads the contents of a file for a given release tag", func() {
			releasesResponse := `
{
	"tag_name": "some-tag",
	"assets": [
		{"name": "some-other-asset-name", "url": "some-other-asset-url"},
		{"name": "some-asset-name", "url": "some-asset-url"}
	]
}
`
			fakeWebClient.GetReturns([]byte(releasesResponse), nil)

			url, err := githubClient.DownloadReleaseAsset("some-org", "some-repo", "some-tag", "some-asset-name", "some-output-path")
			require.NoError(err)

			assert.Equal("some-asset-url", url)

			urlArg, optionsArg := fakeWebClient.GetArgsForCall(0)
			assert.Equal("https://api.github.com/repos/some-org/some-repo/releases/tags/some-tag", urlArg)
			assert.Len(optionsArg, 1)

			request, err := http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))

			urlArg, outputPathArg, optionsArg := fakeWebClient.DownloadArgsForCall(0)
			assert.Equal("some-asset-url", urlArg)
			assert.Len(optionsArg, 2)
			assert.Equal("some-output-path", outputPathArg)

			request, err = http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			optionsArg[1](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))
			assert.Equal("application/octet-stream", request.Header.Get("Accept"))
		})

		when("the asset name does not exist in the release", func() {
			it("returns an AssetNotFound error", func() {
				releasesResponse := `{"tag_name": "some-tag", "assets": []}`
				fakeWebClient.GetReturnsOnCall(0, []byte(releasesResponse), nil)

				_, err := githubClient.DownloadReleaseAsset("some-org", "some-repo", "some-tag", "some-asset-name", "some-output-dir")
				assert.Error(err)
				assert.True(errors.Is(err, internal_errors.AssetNotFound{AssetName: "some-asset-name"}))
			})
		})
	})

	when("DownloadSourceTarball", func() {
		it("downloads the source tarball for a given ref", func() {
			url, err := githubClient.DownloadSourceTarball("some-org", "some-repo", "some-ref", "some-output-path")
			require.NoError(err)

			assert.Equal("https://github.com/some-org/some-repo/tarball/some-ref", url)

			urlArg, outputPathArg, optionsArg := fakeWebClient.DownloadArgsForCall(0)
			assert.Equal(url, urlArg)
			assert.Len(optionsArg, 2)
			assert.Equal("some-output-path", outputPathArg)

			request, err := http.NewRequest("", "", nil)
			require.NoError(err)
			optionsArg[0](request)
			optionsArg[1](request)
			assert.Equal("token some-access-token", request.Header.Get("Authorization"))
			assert.Equal("application/octet-stream", request.Header.Get("Accept"))
		})
	})

	when("GetTagCommit", func() {
		it("returns the commit for the given tag", func() {
			tagResponse := `{"object": {"sha": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}`
			commitResponse := `
{
  "sha": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "commit": {
	"committer": {
	  "date": "2020-01-31T00:00:00Z"
	}
  }
}
`
			fakeWebClient.GetReturnsOnCall(0, []byte(tagResponse), nil)
			fakeWebClient.GetReturnsOnCall(1, []byte(commitResponse), nil)

			actualTagCommit, err := githubClient.GetTagCommit("some-org", "some-repo", "1.0.0")
			require.NoError(err)

			expectedTagCommit := internal.GithubTagCommit{
				Tag:  "1.0.0",
				SHA:  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				Date: "2020-01-31T00:00:00Z",
			}

			assert.Equal(expectedTagCommit, actualTagCommit)
		})
	})
}
