package dependency

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
)

type Composer struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	githubClient     GithubClient
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

func (c Composer) GetAllVersionRefs() ([]string, error) {
	releases, err := c.githubClient.GetReleaseTags("composer", "composer")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
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

func (c Composer) GetDependencyVersion(version string) (DepVersion, error) {
	releases, err := c.githubClient.GetReleaseTags("composer", "composer")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.TagName == version {
			depVersion, err := c.createDependencyVersion(release)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not create composer version: %w", err)
			}
			return depVersion, nil
		}
	}

	return DepVersion{}, fmt.Errorf("could not find composer version %s", version)
}

func (c Composer) GetReleaseDate(version string) (*time.Time, error) {
	releases, err := c.githubClient.GetReleaseTags("composer", "composer")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.TagName == version {
			return &release.PublishedDate, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (c Composer) createDependencyVersion(release internal.GithubRelease) (DepVersion, error) {
	sha, err := c.getDependencySHA(release.TagName)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get sha: %w", err)
	}

	depURL := c.dependencyURL(release.TagName)
	licenses, err := c.licenseRetriever.LookupLicenses("composer", depURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not find license metadata: %w", err)
	}

	return DepVersion{
		Version:         release.TagName,
		URI:             depURL,
		SHA256:          sha,
		ReleaseDate:     &release.PublishedDate,
		DeprecationDate: nil,
		PURL:            c.purlGenerator.Generate("composer", release.TagName, sha, depURL),
		Licenses:        licenses,
	}, nil
}

func (c Composer) getDependencySHA(version string) (string, error) {
	shaUrl := c.shaURL(version)
	body, err := c.webClient.Get(shaUrl)
	if err != nil {
		return "", fmt.Errorf("could not download composer SHA256 file: %w", err)
	}
	depSHA := strings.Split(string(body), " ")[0]
	if len(depSHA) < 64 {
		return "", fmt.Errorf("could not get SHA256 from file %s", shaUrl)
	}
	return depSHA, nil
}

func (c Composer) dependencyURL(version string) string {
	return fmt.Sprintf("https://getcomposer.org/download/%s/composer.phar", version)
}

func (c Composer) shaURL(version string) string {
	return fmt.Sprintf("https://getcomposer.org/download/%s/composer.phar.sha256sum", version)
}
