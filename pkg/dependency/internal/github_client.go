package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
)

type GithubReleaseResponse struct {
	TagName     string               `json:"tag_name"`
	Draft       bool                 `json:"draft"`
	Prerelease  bool                 `json:"prerelease"`
	Assets      []GithubReleaseAsset `json:"assets"`
	PublishedAt time.Time            `json:"published_at"`
	CreatedAt   time.Time            `json:"created_at"`
}

type GithubReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type GithubRelease struct {
	TagName       string
	PublishedDate time.Time
	CreatedDate   time.Time
}

type GithubTagResponse struct {
	Object struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

type GithubTagCommit struct {
	Tag  string
	SHA  string
	Date time.Time
}

type GithubCommitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

type GithubGraphQLRequest struct {
	Query string `json:"query"`
}

type GithubGraphQLTagsResponse struct {
	Data struct {
		Repository struct {
			Refs struct {
				Edges []struct {
					Node struct {
						Name   string
						Target struct {
							OID string `json:"oid"`
						} `json:"target"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"refs"`
		} `json:"repository"`
	} `json:"data"`
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GithubWebClient
type GithubWebClient interface {
	Get(url string, options ...RequestOption) ([]byte, error)
	Post(url string, body []byte, options ...RequestOption) ([]byte, error)
	Download(url, filename string, options ...RequestOption) error
}

type GithubClient struct {
	webClient   GithubWebClient
	accessToken string
}

func NewGithubClient(webClient GithubWebClient, accessToken string) GithubClient {
	return GithubClient{
		webClient:   webClient,
		accessToken: accessToken,
	}
}

func (g GithubClient) GetReleaseTags(org, repo string) ([]GithubRelease, error) {
	page := 1
	var allReleases []GithubRelease
	for {
		body, err := g.webClient.Get(
			fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=100&page=%d", org, repo, page),
			WithHeader("Authorization", "token "+g.accessToken),
		)
		if err != nil {
			return nil, fmt.Errorf("could not get releases: %w", err)
		}

		var releases []GithubReleaseResponse
		err = json.Unmarshal(body, &releases)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal releases: %w\n%s", err, body)
		}

		if len(releases) == 0 {
			break
		}

		page++

		for _, release := range releases {
			if release.Draft || release.Prerelease {
				continue
			}

			allReleases = append(allReleases, GithubRelease{
				TagName:       release.TagName,
				PublishedDate: release.PublishedAt,
				CreatedDate:   release.CreatedAt,
			})
		}
	}

	return allReleases, nil
}

func (g GithubClient) GetTags(org, repo string) ([]string, error) {
	query := fmt.Sprintf(`
	{
		repository(owner: "%s", name: "%s") {
			refs(refPrefix: "refs/tags/", first: 100, orderBy: {field: TAG_COMMIT_DATE, direction: DESC}) {
				edges {
					node {
						name
						target {
							... on Tag {
								tagger {
									name
								}
							}
						}
					}
				}
			}
		}
	}`, org, repo)

	requestBody, err := json.Marshal(GithubGraphQLRequest{Query: query})
	if err != nil {
		return nil, fmt.Errorf("could not marshal graphql request: %w", err)
	}

	body, err := g.webClient.Post(
		"https://api.github.com/graphql",
		requestBody,
		WithHeader("Authorization", "token "+g.accessToken),
	)
	if err != nil {
		return nil, fmt.Errorf("could not make query graphql: %w", err)
	}

	var tagsResponse GithubGraphQLTagsResponse
	err = json.Unmarshal(body, &tagsResponse)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w\n%s", err, body)
	}

	var tags []string
	for _, edge := range tagsResponse.Data.Repository.Refs.Edges {
		tags = append(tags, edge.Node.Name)
	}

	return tags, nil
}

func (g GithubClient) GetReleaseAsset(org, repo, tag, assetName string) ([]byte, error) {
	assetURL, err := g.getReleaseAssetURL(org, repo, tag, assetName)
	if err != nil {
		return nil, err
	}

	assetContents, err := g.webClient.Get(
		assetURL,
		WithHeader("Authorization", "token "+g.accessToken),
		WithHeader("Accept", "application/octet-stream"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not get asset: %w", err)
	}

	return assetContents, nil
}

func (g GithubClient) DownloadReleaseAsset(org, repo, tag, assetName, outputPath string) (string, error) {
	assetURL, err := g.getReleaseAssetURL(org, repo, tag, assetName)
	if err != nil {
		return "", err
	}

	err = g.webClient.Download(
		assetURL,
		outputPath,
		WithHeader("Authorization", "token "+g.accessToken),
		WithHeader("Accept", "application/octet-stream"),
	)
	if err != nil {
		return "", fmt.Errorf("could not get asset: %w", err)
	}

	return assetURL, nil
}

func (g GithubClient) DownloadSourceTarball(org, repo, ref, outputPath string) (string, error) {
	assetURL := fmt.Sprintf("https://github.com/%s/%s/tarball/%s", org, repo, ref)

	err := g.webClient.Download(
		assetURL,
		outputPath,
		WithHeader("Authorization", "token "+g.accessToken),
		WithHeader("Accept", "application/octet-stream"),
	)
	if err != nil {
		return "", fmt.Errorf("could not get tarball: %w", err)
	}

	return assetURL, nil
}

func (g GithubClient) GetTagCommit(org, repo, tag string) (GithubTagCommit, error) {
	body, err := g.webClient.Get(
		fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/tags/%s", org, repo, tag),
		WithHeader("Authorization", "token "+g.accessToken),
	)
	if err != nil {
		return GithubTagCommit{}, fmt.Errorf("could not get tag %s: %w", tag, err)
	}

	var tagResponse GithubTagResponse
	err = json.Unmarshal(body, &tagResponse)
	if err != nil {
		return GithubTagCommit{}, fmt.Errorf("could not unmarshal tag: %w\n%s", err, body)
	}

	body, err = g.webClient.Get(
		fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", org, repo, tagResponse.Object.SHA),
		WithHeader("Authorization", "token "+g.accessToken),
	)
	if err != nil {
		return GithubTagCommit{}, fmt.Errorf("could not get commit: %w", err)
	}

	var githubCommitResponse GithubCommitResponse
	err = json.Unmarshal(body, &githubCommitResponse)
	if err != nil {
		return GithubTagCommit{}, fmt.Errorf("could not unmarshal releases: %w\n%s", err, body)
	}

	return GithubTagCommit{
		Tag:  tag,
		SHA:  githubCommitResponse.SHA,
		Date: githubCommitResponse.Commit.Committer.Date,
	}, nil
}

func (g GithubClient) getReleaseAssetURL(org, repo, tag, assetName string) (string, error) {
	body, err := g.webClient.Get(
		fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", org, repo, tag),
		WithHeader("Authorization", "token "+g.accessToken),
	)
	if err != nil {
		return "", fmt.Errorf("could not get release: %w", err)
	}

	var release GithubReleaseResponse
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal release: %w\n%s", err, body)
	}

	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.URL
			break
		}
	}

	if assetURL == "" {
		return "", internal_errors.AssetNotFound{AssetName: assetName}
	}

	return assetURL, nil
}
