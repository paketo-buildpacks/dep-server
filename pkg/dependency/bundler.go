package dependency

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

type Bundler struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type BundlerRelease struct {
	Version string `json:"number"`
	Date    string `json:"created_at"`
	SHA     string `json:"sha"`
}

func (b Bundler) GetAllVersionRefs() ([]string, error) {
	bundlerReleases, err := b.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get bundler releases: %w", err)
	}

	re := regexp.MustCompile(`\d+\.\d+\.\d+$`)

	var versions []string
	for _, release := range bundlerReleases {
		if !re.MatchString(release.Version) {
			continue
		}
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (b Bundler) GetDependencyVersion(version string) (DepVersion, error) {
	bundlerReleases, err := b.getAllReleases()
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	depURL := b.getDependencyURL(version)

	licenses, err := b.licenseRetriever.LookupLicenses("bundler", depURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	for _, release := range bundlerReleases {
		if release.Version == version {
			releaseDate, err := time.Parse(time.RFC3339Nano, release.Date)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not parse release date: %w", err)
			}
			return DepVersion{
				Version:         version,
				URI:             depURL,
				SHA256:          release.SHA,
				ReleaseDate:     &releaseDate,
				DeprecationDate: nil,
				CPE:             fmt.Sprintf("cpe:2.3:a:bundler:bundler:%s:*:*:*:*:ruby:*:*", version),
				PURL:            b.purlGenerator.Generate("bundler", version, release.SHA, depURL),
				Licenses:        licenses,
			}, nil
		}
	}

	return DepVersion{}, fmt.Errorf("could not find version %s", version)
}

func (b Bundler) GetReleaseDate(version string) (*time.Time, error) {
	bundlerReleases, err := b.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range bundlerReleases {
		if release.Version == version {
			releaseDate, err := time.Parse(time.RFC3339Nano, release.Date)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}
			return &releaseDate, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (b Bundler) getAllReleases() ([]BundlerRelease, error) {
	body, err := b.webClient.Get("https://rubygems.org/api/v1/versions/bundler.json")
	if err != nil {
		return nil, fmt.Errorf("could not get release index: %w", err)
	}

	var bundlerReleases []BundlerRelease
	err = json.Unmarshal(body, &bundlerReleases)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w\n%s", err, body)
	}

	return bundlerReleases, nil
}

func (b Bundler) getDependencyURL(version string) string {
	return fmt.Sprintf("https://rubygems.org/downloads/bundler-%s.gem", version)
}
