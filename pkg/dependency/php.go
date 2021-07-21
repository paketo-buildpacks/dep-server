package dependency

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Php struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type PhpSource struct {
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
	MD5      string `json:"md5"`
}

type PhpRawRelease struct {
	Date   string      `json:"date"`
	Source []PhpSource `json:"source"`
	Museum bool        `json:"museum"`
}

type PhpRelease struct {
	Version string
	Date    *time.Time
	Source  []PhpSource
}

func (p Php) GetAllVersionRefs() ([]string, error) {
	phpReleases, err := p.getPhpReleases()
	if err != nil {
		return nil, err
	}

	sort.Slice(phpReleases, func(i, j int) bool {
		if phpReleases[i].Date.Equal(*phpReleases[j].Date) {
			return phpReleases[i].Version > phpReleases[j].Version
		}
		return phpReleases[i].Date.After(*phpReleases[j].Date)
	})

	var versions []string
	for _, release := range phpReleases {
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (p Php) GetDependencyVersion(version string) (DepVersion, error) {
	release, err := p.getRelease(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release: %w", err)
	}

	dependencyURL := p.dependencyURL(release, version)
	dependencySHA, err := p.getDependencySHA(release, version)
	if err != nil {
		return DepVersion{}, err
	}

	releaseDate, err := p.getReleaseDate(release)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release date: %w", err)
	}

	deprecationDate := p.calculateDeprecationDate(*releaseDate)

	licenses, err := p.licenseRetriever.LookupLicenses("php", dependencyURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             dependencyURL,
		SHA256:          dependencySHA,
		ReleaseDate:     releaseDate,
		DeprecationDate: deprecationDate,
		CPE:             fmt.Sprintf("cpe:2.3:a:php:php:%s:*:*:*:*:*:*:*", version),
		PURL:            p.purlGenerator.Generate("php", version, dependencySHA, dependencyURL),
		Licenses:        licenses,
	}, nil
}

func (p Php) GetReleaseDate(version string) (*time.Time, error) {
	release, err := p.getRelease(version)
	if err != nil {
		return nil, fmt.Errorf("could not get release: %w", err)
	}

	return p.getReleaseDate(release)
}

func (p Php) getPhpReleases() ([]PhpRelease, error) {
	body, err := p.webClient.Get("https://www.php.net/releases/index.php?json")
	if err != nil {
		return nil, fmt.Errorf("could not hit php.net: %w", err)
	}

	var phpLines map[string]interface{}
	err = json.Unmarshal(body, &phpLines)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal version lines response: %w\n%s", err, body)
	}

	var versionLines []string
	for line := range phpLines {
		if line == "3" {
			continue
		}

		versionLines = append(versionLines, line)
	}
	sort.Strings(versionLines)

	var allPhpReleases []PhpRelease

	for _, line := range versionLines {
		body, err = p.webClient.Get(fmt.Sprintf("https://www.php.net/releases/index.php?json&version=%s&max=1000", line))
		if err != nil {
			return nil, fmt.Errorf("could not hit php.net: %w", err)
		}

		var phpRawReleases map[string]PhpRawRelease
		err = json.Unmarshal(body, &phpRawReleases)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal version lines response: %w\n%s", err, body)
		}

		for version, release := range phpRawReleases {
			releaseDate, err := p.parseReleaseDate(release.Date)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}

			allPhpReleases = append(allPhpReleases, PhpRelease{
				Version: version,
				Date:    releaseDate,
				Source:  release.Source,
			})
		}
	}

	return allPhpReleases, nil
}

func (p Php) getRelease(version string) (PhpRawRelease, error) {
	body, err := p.webClient.Get(fmt.Sprintf("https://www.php.net/releases/index.php?json&version=%s", version))
	if err != nil {
		return PhpRawRelease{}, fmt.Errorf("could not hit php.net: %w", err)
	}

	var release PhpRawRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return PhpRawRelease{}, fmt.Errorf("could not unmarshal version response: %w\n%s", err, body)
	}

	return release, nil
}

func (p Php) getDependencySHA(release PhpRawRelease, version string) (string, error) {
	for _, file := range release.Source {
		if filepath.Ext(file.Filename) == ".gz" {
			if file.SHA256 != "" {
				return file.SHA256, nil
			} else if file.MD5 != "" || p.dependencyVersionIsMissingChecksum(version) {
				sha, err := p.getSHA256FromReleaseFile(release, file, version)
				if err != nil {
					return "", fmt.Errorf("could not get SHA256 from release file: %w", err)
				}

				return sha, nil
			} else {
				return "", fmt.Errorf("could not find SHA256 or MD5 for %s", version)
			}
		}
	}

	return "", fmt.Errorf("could not find .tar.gz file for %s", version)
}

func (p Php) getReleaseDate(release PhpRawRelease) (*time.Time, error) {
	if parsedDate, err := time.Parse("02 Jan 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 Jan 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("02 January 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 January 2006", release.Date); err == nil {
		return &parsedDate, nil
	}

	return nil, fmt.Errorf("release date '%s' did not match any expected patterns", release.Date)
}

func (p Php) calculateDeprecationDate(releaseDate time.Time) *time.Time {
	deprecationDate := time.Date(releaseDate.Year()+3, releaseDate.Month(), releaseDate.Day(),
		0, 0, 0, 0, time.UTC)

	return &deprecationDate
}

func (p Php) dependencyURL(release PhpRawRelease, version string) string {
	if release.Museum {
		majorVersion := version[0:1]
		return fmt.Sprintf("https://museum.php.net/php%s/php-%s.tar.gz", majorVersion, version)
	}

	return fmt.Sprintf("https://www.php.net/distributions/php-%s.tar.gz", version)
}

func (p Php) parseReleaseDate(date string) (*time.Time, error) {
	if parsedDate, err := time.Parse("02 Jan 2006", date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 Jan 2006", date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("02 January 2006", date); err == nil {
		return &parsedDate, nil
	}

	if parsedDate, err := time.Parse("2 January 2006", date); err == nil {
		return &parsedDate, nil
	}

	return nil, fmt.Errorf("release date '%s' did not match any expected patterns", date)
}

func (p Php) getSHA256FromReleaseFile(release PhpRawRelease, file PhpSource, version string) (string, error) {
	tempDir, err := ioutil.TempDir("", "php")
	if err != nil {
		return "", fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dependencyOutputPath := filepath.Join(tempDir, file.Filename)
	err = p.webClient.Download(p.dependencyURL(release, version), dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	if !p.dependencyVersionHasIncorrectChecksum(version) && file.MD5 != "" {
		err = p.checksummer.VerifyMD5(dependencyOutputPath, file.MD5)
		if err != nil {
			return "", fmt.Errorf("dependency signature verification failed: %w", err)
		}
	}

	sha256, err := p.checksummer.GetSHA256(dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not get SHA256: %w", err)
	}

	return sha256, nil
}

func (p Php) dependencyVersionIsMissingChecksum(version string) bool {
	versionsWithMissingChecksum := map[string]bool{
		"5.1.6":  true,
		"5.1.5":  true,
		"5.1.4":  true,
		"5.1.3":  true,
		"5.1.2":  true,
		"5.1.1":  true,
		"5.1.0":  true,
		"5.0.5":  true,
		"5.0.4":  true,
		"5.0.3":  true,
		"5.0.2":  true,
		"5.0.1":  true,
		"5.0.0":  true,
		"4.4.5":  true,
		"4.4.4":  true,
		"4.4.3":  true,
		"4.4.2":  true,
		"4.4.1":  true,
		"4.4.0":  true,
		"4.3.11": true,
		"4.3.10": true,
		"4.3.9":  true,
		"4.3.8":  true,
		"4.3.7":  true,
		"4.3.6":  true,
		"4.3.5":  true,
		"4.3.4":  true,
		"4.3.3":  true,
		"4.3.2":  true,
		"4.3.1":  true,
		"4.3.0":  true,
		"4.2.3":  true,
		"4.2.2":  true,
		"4.2.1":  true,
		"4.2.0":  true,
		"4.1.2":  true,
		"4.1.1":  true,
		"4.1.0":  true,
		"4.0.6":  true,
		"4.0.5":  true,
		"4.0.4":  true,
		"4.0.3":  true,
		"4.0.2":  true,
		"4.0.1":  true,
		"4.0.0":  true,
	}

	_, shouldBeIgnored := versionsWithMissingChecksum[version]
	return shouldBeIgnored
}

func (p Php) dependencyVersionHasIncorrectChecksum(version string) bool {
	versionsWithWrongChecksum := map[string]bool{
		"5.3.25": true,
		"5.3.11": true,
		"5.2.14": true,
	}

	_, shouldBeIgnored := versionsWithWrongChecksum[version]
	return shouldBeIgnored
}
