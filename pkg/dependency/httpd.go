package dependency

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type Httpd struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type HttpdRelease struct {
	version       string
	releaseDate   time.Time
	dependencyURL string
	sha256URL     string
	sha1URL       string
	md5URL        string
}

func (h Httpd) GetAllVersionRefs() ([]string, error) {
	releases, err := h.getReleases("")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	err = h.sortReleases(releases)
	if err != nil {
		return nil, fmt.Errorf("could not sort releases: %w", err)
	}

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.version)
	}

	return versions, nil
}

func (h Httpd) GetDependencyVersion(version string) (DepVersion, error) {
	release, err := h.getRelease(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release: %w", err)
	}

	sha, err := h.getDependencySHA256(release)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get sha256 for dependency: %w", err)
	}

	depURL := release.dependencyURL
	licenses, err := h.licenseRetriever.LookupLicenses("httpd", depURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:     version,
		URI:         depURL,
		SHA256:      sha,
		ReleaseDate: &release.releaseDate,
		CPE:         fmt.Sprintf("cpe:2.3:a:apache:http_server:%s:*:*:*:*:*:*:*", version),
		PURL:        h.purlGenerator.Generate("httpd", version, sha, depURL),
		Licenses:    licenses,
	}, nil
}

func (h Httpd) GetReleaseDate(version string) (*time.Time, error) {
	release, err := h.getRelease(version)
	if err != nil {
		return nil, fmt.Errorf("could not get release: %w", err)
	}
	return &release.releaseDate, nil
}

func (h Httpd) getRelease(version string) (HttpdRelease, error) {
	releases, err := h.getReleases(version)
	if err != nil {
		return HttpdRelease{}, fmt.Errorf("could not get releases: %w", err)
	}
	if len(releases) != 1 {
		return HttpdRelease{}, fmt.Errorf("expected to find 1 release but found %d (%v)", len(releases), releases)
	}

	return releases[0], nil
}

func (h Httpd) getReleases(versionFilter string) ([]HttpdRelease, error) {
	filePattern := "httpd-*.tar.bz2*"
	if versionFilter != "" {
		filePattern = fmt.Sprintf("httpd-%s.tar.bz2*", versionFilter)
	}

	body, err := h.webClient.Get("http://archive.apache.org/dist/httpd/?F=2&C=M&O=D&P=" + filePattern)
	if err != nil {
		return nil, fmt.Errorf("could not get file list from archive.apache.org: %w", err)
	}

	re := regexp.MustCompile(`>httpd-([\d\.]+)\.tar\.bz2<.*(\d\d\d\d-\d\d-\d\d \d\d:\d\d)`)

	var releases []HttpdRelease
	for _, line := range strings.Split(string(body), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		version := matches[1]

		date, err := time.Parse("2006-01-02 15:04", matches[2])
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' as date for version '%s'", matches[2], version)
		}

		releases = append(releases, HttpdRelease{
			version:       version,
			releaseDate:   date,
			dependencyURL: h.dependencyURL(version),
			sha256URL:     h.sha256URL(string(body), version),
			sha1URL:       h.sha1URL(string(body), version),
			md5URL:        h.md5URL(string(body), version),
		})
	}

	return releases, nil
}

func (h Httpd) dependencyURL(version string) string {
	return fmt.Sprintf("http://archive.apache.org/dist/httpd/httpd-%s.tar.bz2", version)
}

func (h Httpd) sha256URL(index string, version string) string {
	return h.checksumURL(index, version, "sha256")
}

func (h Httpd) sha1URL(index string, version string) string {
	return h.checksumURL(index, version, "sha1")
}

func (h Httpd) md5URL(index string, version string) string {
	return h.checksumURL(index, version, "md5")
}

func (h Httpd) checksumURL(index string, version string, checksum string) string {
	checksumFilename := fmt.Sprintf("httpd-%s.tar.bz2.%s", version, checksum)
	if strings.Contains(index, checksumFilename) {
		return fmt.Sprintf("http://archive.apache.org/dist/httpd/%s", checksumFilename)
	}
	return ""
}

func (h Httpd) sortReleases(releases []HttpdRelease) error {
	var sortErr error

	sort.Slice(releases, func(i, j int) bool {
		if releases[i].releaseDate != releases[j].releaseDate {
			return releases[i].releaseDate.After(releases[j].releaseDate)
		}

		var v1, v2 *semver.Version

		v1, sortErr = semver.NewVersion(releases[i].version)
		if sortErr != nil {
			return false
		}

		v2, sortErr = semver.NewVersion(releases[j].version)
		if sortErr != nil {
			return false
		}

		return v1.GreaterThan(v2)
	})

	return sortErr
}

func (h Httpd) getDependencySHA256(release HttpdRelease) (string, error) {
	if release.sha256URL == "" && release.sha1URL == "" && release.md5URL == "" && !h.dependencyVersionIsMissingChecksum(release.version) {
		return "", errors.New("could not find checksum file")
	}

	if release.sha256URL != "" {
		checksumContents, err := h.webClient.Get(release.sha256URL)
		if err != nil {
			return "", fmt.Errorf("could not download sha256 file: %w", err)
		}

		return strings.Fields(string(checksumContents))[0], nil
	}

	tempDir, err := ioutil.TempDir("", "httpd")
	if err != nil {
		return "", fmt.Errorf("could not make temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dependencyPath := filepath.Join(tempDir, filepath.Base(release.dependencyURL))

	err = h.webClient.Download(release.dependencyURL, dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	err = h.verifyChecksum(release, dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not verify checksum: %w", err)
	}

	sha256, err := h.checksummer.GetSHA256(dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not get sha256: %w", err)
	}

	return sha256, nil
}

func (h Httpd) verifyChecksum(release HttpdRelease, dependencyPath string) error {
	if h.dependencyVersionIsMissingChecksum(release.version) {
		return nil
	}

	if release.sha1URL != "" {
		checksumContents, err := h.webClient.Get(release.sha1URL)
		if err != nil {
			return fmt.Errorf("could not download sha1 file: %w", err)
		}

		fields := strings.Fields(string(checksumContents))

		var checksum string
		if strings.HasPrefix(fields[0], "SHA1") {
			checksum = fields[len(fields)-1]
		} else {
			checksum = fields[0]
		}

		err = h.checksummer.VerifySHA1(dependencyPath, checksum)
		if err != nil {
			return fmt.Errorf("could not verify sha1: %w", err)
		}
	} else if release.md5URL != "" {
		checksumContents, err := h.webClient.Get(release.md5URL)
		if err != nil {
			return fmt.Errorf("could not download md5 file: %w", err)
		}

		fields := strings.Fields(string(checksumContents))

		var checksum string
		if strings.HasPrefix(fields[0], "MD5") {
			checksum = fields[len(fields)-1]
		} else {
			checksum = fields[0]
		}

		err = h.checksummer.VerifyMD5(dependencyPath, checksum)
		if err != nil {
			return fmt.Errorf("could not verify md5: %w", err)
		}
	}

	return nil
}

func (h Httpd) dependencyVersionIsMissingChecksum(version string) bool {
	versionsWithMissingChecksum := map[string]bool{
		"2.2.3": true,
	}

	_, shouldBeIgnored := versionsWithMissingChecksum[version]
	return shouldBeIgnored
}
