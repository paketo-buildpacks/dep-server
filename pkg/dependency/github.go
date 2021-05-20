package dependency

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
)

type Github struct {
	productName  string
	checksummer  Checksummer
	fileSystem   FileSystem
	githubClient GithubClient
	webClient    WebClient
}

func (g Github) GetAllVersionRefs() ([]string, error) {
	splitName := strings.Split(g.productName, "/")
	org, repo := splitName[0], splitName[1]
	releases, err := g.githubClient.GetReleaseTags(org, repo)
	if err != nil {
		// return nil, fmt.Errorf("could not get releases: %w", err)
		panic(err)
	}

	var versions []string
	for _, release := range releases {
		version, err := semver.NewVersion(release.TagName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version: %w", err)
		}
		if version.Prerelease() != "" {
			continue
		}
		versions = append(versions, release.TagName)
	}
	return versions, nil
}

func (g Github) GetDependencyVersion(version string) (DepVersion, error) {
	splitName := strings.Split(g.productName, "/")
	org, repo := splitName[0], splitName[1]
	releases, err := g.githubClient.GetReleaseTags(org, repo)
	if err != nil {
		// return nil, fmt.Errorf("could not get releases: %w", err)
		panic(err)
	}

	for _, release := range releases {
		if release.TagName == version {
			depVersion, err := g.createDependencyVersion(release)
			if err != nil {
				panic(err)
			}
			return depVersion, nil
		}
	}

	return DepVersion{}, fmt.Errorf("could not find %s version %s", repo, version)
}

func (g Github) GetReleaseDate(version string) (*time.Time, error) {
	splitName := strings.Split(g.productName, "/")
	org, repo := splitName[0], splitName[1]
	releases, err := g.githubClient.GetReleaseTags(org, repo)
	if err != nil {
		// return nil, fmt.Errorf("could not get releases: %w", err)
		panic(err)
	}

	for _, release := range releases {
		if release.TagName == version {
			return &release.PublishedDate, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (g Github) dependencyURL(name, version string) string {
	return fmt.Sprintf("https://github.com/%s/archive/v%s.tar.gz", name, version)
}
func (g Github) createDependencyVersion(release internal.GithubRelease) (DepVersion, error) {
	return DepVersion{
		Version:         release.TagName,
		URI:             g.dependencyURL(g.productName, release.TagName),
		ReleaseDate:     &release.PublishedDate,
		DeprecationDate: nil,
	}, nil
}
