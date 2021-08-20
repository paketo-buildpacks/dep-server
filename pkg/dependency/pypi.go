package dependency

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type PyPi struct {
	productName string

	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type PyPiRelease struct {
	Version    string
	URL        string
	SHA256     string
	UploadTime time.Time
}

func (p PyPi) GetAllVersionRefs() ([]string, error) {
	releases, err := p.getReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	err = p.sortReleases(releases)
	if err != nil {
		return nil, fmt.Errorf("could not sort releases: %w", err)
	}

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (p PyPi) GetDependencyVersion(version string) (DepVersion, error) {
	releases, err := p.getReleases()
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.Version == version {
			return release, nil
		}
	}

	return DepVersion{}, fmt.Errorf("could not find release with version %s: %w", version, err)
}

func (p PyPi) GetReleaseDate(version string) (*time.Time, error) {
	releases, err := p.getReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.Version == version {
			return release.ReleaseDate, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (p PyPi) getReleases() ([]DepVersion, error) {
	body, err := p.webClient.Get(fmt.Sprintf("https://pypi.org/pypi/%s/json", p.productName))
	if err != nil {
		return nil, fmt.Errorf("could not get project metadata: %w", err)
	}

	var productMetadata struct {
		Releases map[string][]struct {
			PackageType string            `json:"packagetype"`
			URL         string            `json:"url"`
			UploadTime  string            `json:"upload_time_iso_8601"`
			Digests     map[string]string `json:"digests"`
		} `json:"releases"`
	}
	err = json.Unmarshal(body, &productMetadata)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal project metadata: %w", err)
	}

	var releases []DepVersion
	for version, releasesForVersion := range productMetadata.Releases {
		for _, release := range releasesForVersion {
			if strings.Contains(version, "b") || strings.Contains(version, "dev") {
				continue
			}
			if release.PackageType != "sdist" {
				continue
			}

			uploadTime, err := time.Parse(time.RFC3339, release.UploadTime)
			if err != nil {
				return nil, fmt.Errorf("could not parse upload time '%s' as date for version %s: %w", release.UploadTime, version, err)
			}

			if release.Digests["sha256"] == "" {
				return nil, fmt.Errorf("could not find sha256 for version %s", version)
			}

			var cpe string
			if p.productName == "pip" {
				cpe = fmt.Sprintf("cpe:2.3:a:pypa:pip:%s:*:*:*:*:python:*:*", version)
			}

			licenses, err := p.licenseRetriever.LookupLicenses("pypi", release.URL)
			if err != nil {
				return nil, fmt.Errorf("could not get retrieve licenses: %w", err)
			}

			releases = append(releases, DepVersion{
				Version:     version,
				URI:         release.URL,
				SHA256:      release.Digests["sha256"],
				ReleaseDate: &uploadTime,
				CPE:         cpe,
				PURL:        p.purlGenerator.Generate(p.productName, version, release.Digests["sha256"], release.URL),
				Licenses:    licenses,
			})
		}
	}

	return releases, nil
}

func (p PyPi) sortReleases(releases []DepVersion) error {
	var sortErr error
	sort.Slice(releases, func(i, j int) bool {
		if *releases[i].ReleaseDate != *releases[j].ReleaseDate {
			return releases[i].ReleaseDate.After(*releases[j].ReleaseDate)
		}

		semver1, err := semver.NewVersion(releases[i].Version)
		if err != nil {
			sortErr = fmt.Errorf("could not parse '%s' as semver", releases[i].Version)
			return false
		}
		semver2, err := semver.NewVersion(releases[j].Version)
		if err != nil {
			sortErr = fmt.Errorf("could not parse '%s' as semver", releases[j].Version)
			return false
		}
		return semver1.GreaterThan(semver2)
	})
	return sortErr
}
