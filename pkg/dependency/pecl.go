package dependency

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type Pecl struct {
	productName string

	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type PeclVersion struct {
	Name        string
	Version     string
	ReleaseDate *time.Time
}

func (p Pecl) GetAllVersionRefs() ([]string, error) {
	peclVersions, err := p.getVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency versions: %w", err)
	}

	var versions []string
	for _, peclVersion := range peclVersions {
		versions = append(versions, peclVersion.Version)
	}
	return versions, nil
}

func (p Pecl) GetDependencyVersion(version string) (DepVersion, error) {
	versions, err := p.getVersions()
	if err != nil {
		return DepVersion{}, fmt.Errorf("failed to get dependency versions: %w", err)
	}

	for _, currVersion := range versions {
		if currVersion.Version == version {
			dependencyURL := fmt.Sprintf("https://pecl.php.net/get/%s-%s", currVersion.Name, currVersion.Version)

			dependencyOutputDir, err := ioutil.TempDir("", "pecl")
			if err != nil {
				return DepVersion{}, fmt.Errorf("failed to create temp dir: %w", err)
			}

			dependencyOutputPath := filepath.Join(dependencyOutputDir, filepath.Base(dependencyURL))
			err = p.webClient.Download(dependencyURL, dependencyOutputPath)
			if err != nil {
				return DepVersion{}, fmt.Errorf("failed to download dependency at %s: %w", dependencyURL, err)
			}

			dependencySHA, err := p.checksummer.GetSHA256(dependencyOutputPath)
			if err != nil {
				return DepVersion{}, fmt.Errorf("failed to generate dependency checksum: %w", err)
			}

			licenses, err := p.licenseRetriever.LookupLicenses("pecl", dependencyURL)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
			}

			return DepVersion{
				Version:         currVersion.Version,
				URI:             dependencyURL,
				SHA256:          dependencySHA,
				ReleaseDate:     currVersion.ReleaseDate,
				DeprecationDate: nil,
				PURL:            p.purlGenerator.Generate(currVersion.Name, version, dependencySHA, dependencyURL),
				Licenses:        licenses,
			}, nil
		}
	}
	return DepVersion{}, nil
}

func (p Pecl) GetReleaseDate(version string) (*time.Time, error) {
	versions, err := p.getVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency versions: %w", err)
	}

	for _, currVersion := range versions {
		if currVersion.Version == version {
			return currVersion.ReleaseDate, nil
		}
	}
	return nil, nil
}

func (p Pecl) getVersions() ([]PeclVersion, error) {
	body, err := p.webClient.Get(fmt.Sprintf("https://pecl.php.net/feeds/pkg_%s.rss", p.productName))
	if err != nil {
		return nil, fmt.Errorf("could not get rss feed: %w", err)
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(body))
	if err != nil {
		return nil, fmt.Errorf("error parsing rss feed: %w", err)
	}

	var feedVersions []PeclVersion

	re := regexp.MustCompile(`\d+\.\d+(\.\d+)?$`)
	for _, item := range feed.Items {
		splitTitle := strings.Split(item.Title, " ")

		if re.MatchString(splitTitle[1]) {
			feedVersions = append(feedVersions, PeclVersion{
				Name:        splitTitle[0],
				Version:     splitTitle[1],
				ReleaseDate: item.PublishedParsed,
			})
		}

	}
	return feedVersions, nil
}
