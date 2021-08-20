package dependency

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"

	"github.com/Masterminds/semver"
)

type Curl struct {
	checksummer      Checksummer
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type CurlRelease struct {
	Version string
	Date    time.Time
	Semver  semver.Version
}

const (
	CurlVersionIndex = 1
	CurlDateIndex    = 3
)

func (c Curl) GetAllVersionRefs() ([]string, error) {
	curlReleases, err := c.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get curl releases: %w", err)
	}

	sort.SliceStable(curlReleases, func(i, j int) bool {
		return curlReleases[i].Date.After(curlReleases[j].Date)
	})

	var versions []string
	for _, release := range curlReleases {
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (c Curl) GetDependencyVersion(version string) (DepVersion, error) {
	curlReleases, err := c.getAllReleases()
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range curlReleases {
		if release.Version == version {
			return c.createDependencyVersion(release)
		}
	}

	return DepVersion{}, fmt.Errorf("could not find version %s", version)
}

func (c Curl) GetReleaseDate(version string) (*time.Time, error) {
	curlReleases, err := c.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range curlReleases {
		if release.Version == version {
			return &release.Date, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (c Curl) getAllReleases() ([]CurlRelease, error) {
	body, err := c.webClient.Get("https://curl.se/docs/releases.csv")
	if err != nil {
		return nil, fmt.Errorf("could not get release csv: %w", err)
	}

	var curlReleases []CurlRelease
	r := csv.NewReader(bytes.NewReader(body))
	r.Comma = ';'
	for {
		release, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not read from csv reader: %w", err)
		}

		version, err := semver.NewVersion(release[CurlVersionIndex])
		if err != nil {
			return nil, fmt.Errorf("could not parse version: %w", err)
		}

		if !c.versionHasDownload(*version) {
			continue
		}

		date, err := time.Parse("2006-01-02", release[CurlDateIndex])
		if err != nil {
			return nil, fmt.Errorf("could not parse date: %w", err)
		}

		curlReleases = append(curlReleases, CurlRelease{
			Version: release[CurlVersionIndex],
			Semver:  *version,
			Date:    date,
		})
	}

	return curlReleases, nil
}

func (c Curl) createDependencyVersion(release CurlRelease) (DepVersion, error) {
	sha, err := c.getDependencySHA(release)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get curl sha: %w", err)
	}

	depURL := c.dependencyURL(release)
	licenses, err := c.licenseRetriever.LookupLicenses("curl", depURL)

	return DepVersion{
		Version:     release.Version,
		URI:         depURL,
		SHA256:      sha,
		ReleaseDate: &release.Date,
		CPE:         fmt.Sprintf("cpe:2.3:a:haxx:curl:%s:*:*:*:*:*:*:*", release.Version),
		PURL:        c.purlGenerator.Generate("curl", release.Version, sha, depURL),
		Licenses:    licenses,
	}, nil
}

func (c Curl) getDependencySHA(release CurlRelease) (string, error) {
	dependencyURL := c.dependencyURL(release)
	dependencyOutputDir, err := ioutil.TempDir("", "curl")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	dependencyOutputPath := filepath.Join(dependencyOutputDir, filepath.Base(dependencyURL))

	err = c.webClient.Download(dependencyURL, dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	if c.hasSignatureFile(release) {
		curlGPGKey, err := c.webClient.Get("https://daniel.haxx.se/mykey.asc")
		if err != nil {
			return "", fmt.Errorf("could not get curl GPG key: %w", err)
		}

		dependencySignature, err := c.webClient.Get(c.dependencySignatureURL(release.Version))
		if err != nil {
			return "", fmt.Errorf("could not get dependency signature: %w", err)
		}

		err = c.checksummer.VerifyASC(string(dependencySignature), dependencyOutputPath, string(curlGPGKey))
		if err != nil {
			return "", fmt.Errorf("dependency signature verification failed: %w", err)
		}
	}

	dependencySHA, err := c.checksummer.GetSHA256(dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not get SHA256: %w", err)
	}

	return dependencySHA, nil
}

func (c Curl) hasSignatureFile(release CurlRelease) bool {
	return release.Semver.GreaterThan(semver.MustParse("7.29.0"))
}

func (c Curl) dependencySignatureURL(version string) string {
	return fmt.Sprintf("https://curl.se/download/curl-%s.tar.gz.asc", version)
}

func (c Curl) dependencyURL(release CurlRelease) string {
	if release.Semver.LessThan(semver.MustParse("7.30.0")) {
		return fmt.Sprintf("https://curl.se/download/archeology/curl-%s.tar.gz", release.Version)
	}

	return fmt.Sprintf("https://curl.se/download/curl-%s.tar.gz", release.Version)
}

func (c Curl) versionHasDownload(version semver.Version) bool {
	missingDownload := (version.Major() == 4 && !version.Equal(semver.MustParse("4.8"))) ||
		(version.Major() == 5 && !version.Equal(semver.MustParse("5.9"))) ||
		version.Equal(semver.MustParse("6.3")) ||
		version.Equal(semver.MustParse("6.5")) ||
		version.Equal(semver.MustParse("6.5.1")) ||
		version.Equal(semver.MustParse("7.1"))

	return !missingDownload
}
