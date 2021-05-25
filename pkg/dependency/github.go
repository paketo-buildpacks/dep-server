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
	tags, err := g.githubClient.GetTags(org, repo)
	if err != nil {
		// return nil, fmt.Errorf("could not get releases: %w", err)
		panic(err)
	}

	var versions []string
	for _, tag := range tags {
		version, err := semver.NewVersion(tag)
		if err != nil {
			// if err == semver.ErrInvalidSemVer {
			// 	continue
			// }
			return nil, fmt.Errorf("failed to parse version %s: %w", tag, err)
		}
		if version.Prerelease() != "" {
			continue
		}
		versions = append(versions, tag)
	}
	return versions, nil
}

func (g Github) GetDependencyVersion(version string) (DepVersion, error) {

	releaseDate, err := g.GetReleaseDate(version)
	if err != nil {
		return DepVersion{}, err
	}

	dependencyURL := g.dependencyURL(g.productName, version)

	return DepVersion{
		Version:     version,
		URI:         dependencyURL,
		ReleaseDate: releaseDate,
	}, nil
}

func (g Github) GetReleaseDate(version string) (*time.Time, error) {
	splitName := strings.Split(g.productName, "/")
	org, repo := splitName[0], splitName[1]
	tagCommit, err := g.githubClient.GetTagCommit(org, repo, version)
	if err != nil {
		// return nil, fmt.Errorf("could not get releases: %w", err)
		panic(err)
	}

	return &tagCommit.Date, nil
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
