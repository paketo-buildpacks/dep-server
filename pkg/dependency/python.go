package dependency

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Python struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

func (p Python) GetAllVersionRefs() ([]string, error) {
	body, err := p.webClient.Get("https://www.python.org/downloads/")
	if err != nil {
		return nil, fmt.Errorf("could not get python downloads: %w", err)
	}

	re := regexp.MustCompile(`release-number.*Python ([\d]+\.[\d]+\.[\d]+)`)

	var versions []string
	for _, line := range strings.Split(string(body), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			versions = append(versions, matches[1])
		}
	}

	return versions, nil
}

func (p Python) GetDependencyVersion(version string) (DepVersion, error) {
	sourceURI, releaseDate, potentialMD5s, err := p.getReleaseMetadata(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release metadata: %w", err)
	}

	sha256, err := p.getDependencySHA256(sourceURI, potentialMD5s, version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get dependency SHA256: %w", err)
	}

	deprecationDate, err := p.getReleaseDeprecationDate(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release deprecation date: %w", err)
	}

	licenses, err := p.licenseRetriever.LookupLicenses("python", sourceURI)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             sourceURI,
		SHA256:          sha256,
		ReleaseDate:     releaseDate,
		DeprecationDate: deprecationDate,
		CPE:             fmt.Sprintf("cpe:2.3:a:python:python:%s:*:*:*:*:*:*:*", version),
		PURL:            p.purlGenerator.Generate("python", version, sha256, sourceURI),
		Licenses:        licenses,
	}, nil
}

func (p Python) GetReleaseDate(version string) (*time.Time, error) {
	_, releaseDate, _, err := p.getReleaseMetadata(version)
	if err != nil {
		return nil, fmt.Errorf("could not get release metadata: %w", err)
	}
	return releaseDate, nil
}

func (p Python) getReleaseMetadata(version string) (string, *time.Time, []string, error) {
	versionWithoutPeriods := strings.ReplaceAll(version, ".", "")
	body, err := p.webClient.Get(fmt.Sprintf("https://www.python.org/downloads/release/python-%s/", versionWithoutPeriods))
	if err != nil {
		return "", nil, nil, fmt.Errorf("could not get python downloads: %w", err)
	}

	releaseDateRegexp := regexp.MustCompile(`Release Date:</strong> ([\w]{3})[\w\.]* ([\d]+, [\d]+)`)
	sourceURIRegexp := regexp.MustCompile(`<a href="(.*)">Gzipped source tar ?ball`)
	md5FromFilesRegexp := regexp.MustCompile(`<td>([0-9a-f]{32})</td>`)
	md5FromPreRegexp := regexp.MustCompile(`([0-9a-f]{32}).*[\d]+.*\.tgz`)
	md5FromBlockQuoteRegexp := regexp.MustCompile(`<tt .*>([0-9a-f]{32})</tt>.*\.tgz`)

	var releaseDateDayAndYear, releaseDateMonth, sourceURI string
	var potentialMD5s []string

	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		releaseDateMatches := releaseDateRegexp.FindStringSubmatch(line)
		if len(releaseDateMatches) == 3 {
			releaseDateMonth = releaseDateMatches[1]
			releaseDateDayAndYear = releaseDateMatches[2]

			continue
		}

		sourceURIMatches := sourceURIRegexp.FindStringSubmatch(line)
		if len(sourceURIMatches) == 2 {
			sourceURI = sourceURIMatches[1]

			md5FromFilesMatches := md5FromFilesRegexp.FindStringSubmatch(lines[i+3])
			if len(md5FromFilesMatches) == 2 {
				potentialMD5s = append(potentialMD5s, md5FromFilesMatches[1])
			}

			continue
		}

		md5FromPreMatches := md5FromPreRegexp.FindStringSubmatch(line)
		if len(md5FromPreMatches) == 2 {
			potentialMD5s = append(potentialMD5s, md5FromPreMatches[1])

			continue
		}

		md5FromBlockQuoteMatches := md5FromBlockQuoteRegexp.FindStringSubmatch(line)
		if len(md5FromBlockQuoteMatches) == 2 {
			potentialMD5s = append(potentialMD5s, md5FromBlockQuoteMatches[1])

			continue
		}
	}

	if sourceURI == "" {
		return "", nil, nil, errors.New("could not find source URI on download page")
	}
	if len(potentialMD5s) == 0 {
		return "", nil, nil, errors.New("could not find MD5 on download page")
	}
	if releaseDateDayAndYear == "" || releaseDateMonth == "" {
		return "", nil, nil, errors.New("could not find release date on download page")
	}

	releaseDate, err := time.Parse("Jan 2, 2006", fmt.Sprintf("%s %s", releaseDateMonth, releaseDateDayAndYear))
	if err != nil {
		return "", nil, nil, fmt.Errorf("could not parse release date: %w", err)
	}

	return sourceURI, &releaseDate, potentialMD5s, nil
}

func (p Python) getDependencySHA256(sourceURI string, potentialMD5s []string, version string) (string, error) {
	tempDir, err := ioutil.TempDir("", "python")
	if err != nil {
		return "", fmt.Errorf("could not make temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dependencyPath := filepath.Join(tempDir, filepath.Base(sourceURI))

	err = p.webClient.Download(sourceURI, dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	if !p.dependencyVersionHasWrongChecksum(version) {
		verifiedMD5 := false
		for _, md5 := range potentialMD5s {
			if err = p.checksummer.VerifyMD5(dependencyPath, md5); err == nil {
				verifiedMD5 = true
				break
			}
		}

		if !verifiedMD5 {
			return "", fmt.Errorf("md5 did not match any of [%s]", strings.Join(potentialMD5s, ","))
		}
	}

	sha256, err := p.checksummer.GetSHA256(dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not get sha256: %w", err)
	}

	return sha256, nil
}

func (p Python) getReleaseDeprecationDate(version string) (*time.Time, error) {
	body, err := p.webClient.Get("https://www.python.org/downloads/")
	if err != nil {
		return nil, fmt.Errorf("could not get python downloads: %w", err)
	}

	versionLine := strings.Join(strings.Split(version, ".")[0:2], ".")
	releaseVersionMatcher := fmt.Sprintf(`release-version">%s`, versionLine)

	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		if strings.Contains(line, releaseVersionMatcher) {
			fullDateRegexp := regexp.MustCompile(`release-end">([\d]{4}-[\d]{2}-[\d]{2})`)
			matches := fullDateRegexp.FindStringSubmatch(lines[i+3])
			if len(matches) == 2 {
				deprecationDate, err := time.Parse("2006-01-02", matches[1])
				if err != nil {
					return nil, fmt.Errorf("could not parse deprecation date: %w", err)
				}

				return &deprecationDate, nil
			}

			dateWithoutDayRegexp := regexp.MustCompile(`release-end">([\d]{4}-[\d]{2})`)
			matches = dateWithoutDayRegexp.FindStringSubmatch(lines[i+3])
			if len(matches) == 2 {
				deprecationDate, err := time.Parse("2006-01", matches[1])
				if err != nil {
					return nil, fmt.Errorf("could not parse deprecation date: %w", err)
				}

				return &deprecationDate, nil
			}
		}
	}

	return nil, nil
}

func (p Python) dependencyVersionHasWrongChecksum(version string) bool {
	versionsWithMissingChecksum := map[string]bool{
		"3.1.0": true,
	}

	_, shouldBeIgnored := versionsWithMissingChecksum[version]
	return shouldBeIgnored
}
