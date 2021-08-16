package dependency

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
)

type Tini struct {
	checksummer      Checksummer
	githubClient     GithubClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
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
		// We expect an official Tini version (ex. vX.X.X) to be passed in.
		// Safeguard against user passing in a non-official semveric version
		version = convertToTiniVersion(version)

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

func (t Tini) GetReleaseDate(version string) (*time.Time, error) {
	releases, err := t.githubClient.GetReleaseTags("krallin", "tini")
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

func (t Tini) createDependencyVersion(version string, release internal.GithubRelease) (DepVersion, error) {
	// this function receives an official Tini version from
	// GetAllDependencies usually, but if a user passes in a semantic version
	// via a workflow this will convert the version to an official Tini version.
	version = convertToTiniVersion(version)

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

	licenses, err := t.licenseRetriever.LookupLicenses("tini", tarballURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	semVersion, err := convertToSemVer(version)
	if err != nil {
		return DepVersion{}, err
	}

	return DepVersion{
		Version:         semVersion,
		URI:             tarballURL,
		SHA256:          dependencySHA,
		ReleaseDate:     &release.PublishedDate,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf("cpe:2.3:a:tini_project:tini:%s:*:*:*:*:*:*:*", strings.TrimPrefix(version, "v")),
		PURL:            t.purlGenerator.Generate("tini", version, dependencySHA, tarballURL),
		Licenses:        licenses,
	}, nil
}

// If a user passes in the semantic version rather than the official vX.X.X
// version, the GetDependencyVersion function will fail since we can't easily
// query the Tini release pages
func convertToTiniVersion(version string) string {
	// add a `v` prefix
	if !strings.Contains(version, "v") {
		version = "v" + version
	}
	return version
}
