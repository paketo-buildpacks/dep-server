package dependency

import (
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type CAAPM struct {
	checksummer Checksummer
	fileSystem  FileSystem
	webClient   WebClient
}

func (c CAAPM) GetAllVersionRefs() ([]string, error) {
	body, err := c.webClient.Get("https://ca.bintray.com/apm-agents/")
	if err != nil {
		return nil, fmt.Errorf("could not hit ca.bintray.com: %w", err)
	}

	re := regexp.MustCompile(`<a href="CA-APM-PHPAgent-(.*)_linux\.tar\.gz">`)
	matches := re.FindAllSubmatch(body, -1)

	var versions []string
	for _, match := range matches {
		versions = append(versions, string(match[1]))
	}

	err = c.sortVersions(versions)
	if err != nil {
		return nil, fmt.Errorf("could not sort versions: %w", err)
	}

	return versions, nil
}

func (c CAAPM) GetDependencyVersion(version string) (DepVersion, error) {
	dependencyURL := c.dependencyURL(version)

	dependencyOutputDir, err := ioutil.TempDir("", "caapm")
	if err != nil {
		return DepVersion{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	dependencyOutputPath := filepath.Join(dependencyOutputDir, filepath.Base(dependencyURL))

	err = c.webClient.Download(dependencyURL, dependencyOutputPath)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not download dependency: %w", err)
	}

	dependencySHA, err := c.checksummer.GetSHA256(dependencyOutputPath)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get SHA256: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             dependencyURL,
		SHA:             dependencySHA,
		ReleaseDate:     "",
		DeprecationDate: "",
	}, nil
}

func (c CAAPM) GetReleaseDate(version string) (time.Time, error) {
	return time.Time{}, fmt.Errorf("cannot determine release dates for CAAPM")
}

func (c CAAPM) dependencyURL(version string) string {
	return fmt.Sprintf("https://ca.bintray.com/apm-agents/CA-APM-PHPAgent-%s_linux.tar.gz", version)
}

func (c CAAPM) sortVersions(versions []string) error {
	var sortErr error

	sort.Slice(versions, func(i, j int) bool {
		var v1, v2 *semver.Version

		v1, sortErr = semver.NewVersion(versions[i])
		if sortErr != nil {
			return false
		}

		v2, sortErr = semver.NewVersion(versions[j])
		if sortErr != nil {
			return false
		}

		return v1.GreaterThan(v2)
	})

	return sortErr
}
