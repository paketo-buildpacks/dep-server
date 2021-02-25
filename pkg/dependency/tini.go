package dependency

import (
	"fmt"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Tini struct {
	checksummer  Checksummer
	githubClient GithubClient
}

func (t Tini) GetAllVersionRefs() ([]string, error) {
	releases, err := t.githubClient.GetReleaseTags("krallin", "tini")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.TagName)
	}
	return versions, nil
}

func (t Tini) GetDependencyVersion(version string) (DepVersion, error) {
	releases, err := t.githubClient.GetReleaseTags("krallin", "tini")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.TagName == version {
			depVersion, err := t.createDependencyVersion(version, release)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not create tini version: %w", err)
			}
			return depVersion, nil
		}
	}
	return DepVersion{}, fmt.Errorf("could not find tini version %s", version)
}

func (t Tini) GetReleaseDate(version string) (time.Time, error) {
	releases, err := t.githubClient.GetReleaseTags("krallin", "tini")
	if err != nil {
		return time.Time{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.TagName == version {
			releaseDate, err := time.Parse(time.RFC3339, release.PublishedDate)
			if err != nil {
				return time.Time{}, fmt.Errorf("could not parse release date: %w", err)
			}
			return releaseDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not find release date for version %s", version)
}

func (t Tini) createDependencyVersion(version string, release internal.GithubRelease) (DepVersion, error) {
	tarballDir, err := ioutil.TempDir("", "tini")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tarballDir)

	tarballPath := filepath.Join(tarballDir, fmt.Sprintf("tini-%s.tar.gz", version))

	tarballURL, err := t.githubClient.DownloadSourceTarball("krallin", "tini", version, tarballPath)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not download source tarball: %w", err)
	}

	dependencySHA, err := t.checksummer.GetSHA256(tarballPath)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get SHA256: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             tarballURL,
		SHA:             dependencySHA,
		ReleaseDate:     release.PublishedDate,
		DeprecationDate: "",
		CPE:             fmt.Sprintf("cpe:2.3:a:tini_project:tini:%s:*:*:*:*:*:*:*", strings.TrimPrefix(version, "v")),
	}, nil
}
