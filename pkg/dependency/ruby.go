package dependency

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
)

type Ruby struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type RubyRelease struct {
	Version string
	Date    string
}

func (r Ruby) GetAllVersionRefs() ([]string, error) {
	rubyReleases, err := r.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get ruby releases: %w", err)
	}

	var versions []string
	for _, release := range rubyReleases {
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (r Ruby) GetDependencyVersion(version string) (DepVersion, error) {
	rubyReleases, err := r.getAllReleases()
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	depURL, depSHA, err := r.getDependencyURLAndSHA(version)
	if err != nil {
		return DepVersion{}, err
	}

	licenses, err := r.licenseRetriever.LookupLicenses("ruby", depURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	for _, release := range rubyReleases {
		if release.Version == version {
			releaseDate, err := time.Parse("2006-01-02", release.Date)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not parse release date: %w", err)
			}

			return DepVersion{
				Version:         version,
				URI:             depURL,
				SHA256:          depSHA,
				ReleaseDate:     &releaseDate,
				DeprecationDate: nil,
				CPE:             fmt.Sprintf("cpe:2.3:a:ruby-lang:ruby:%s:*:*:*:*:*:*:*", version),
				PURL:            r.purlGenerator.Generate("ruby", version, depSHA, depURL),
				Licenses:        licenses,
			}, nil
		}
	}

	return DepVersion{}, fmt.Errorf("could not find version %s", version)
}

func (r Ruby) GetReleaseDate(version string) (*time.Time, error) {
	rubyReleases, err := r.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range rubyReleases {
		if release.Version == version {
			releaseDate, err := time.Parse("2006-01-02", release.Date)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}
			return &releaseDate, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (r Ruby) getAllReleases() ([]RubyRelease, error) {
	body, err := r.webClient.Get("https://www.ruby-lang.org/en/downloads/releases/")
	if err != nil {
		return nil, fmt.Errorf("could not get release index: %w", err)
	}

	re := regexp.MustCompile(`>Ruby (\d+\.\d+\.\d+)</td>\n<td>(\d\d\d\d-\d\d-\d\d)<`)
	versions := re.FindAllStringSubmatch(string(body), -1)

	var rubyReleases []RubyRelease
	for _, version := range versions {
		rubyReleases = append(rubyReleases, RubyRelease{
			Version: version[1],
			Date:    version[2],
		})
	}

	return rubyReleases, nil
}

func (r Ruby) getDependencyURLAndSHA(version string) (string, string, error) {
	URL, SHA, err := r.getDependencyURLAndSHAFromGithub(version)
	if err != nil {
		if errors.Is(err, depErrors.NoSourceCodeError{Version: version}) {
			return r.getDependencyURLAndSHAFromMirror(version)
		} else {
			return "", "", err
		}
	}
	return URL, SHA, nil
}

func (r Ruby) getDependencyURLAndSHAFromGithub(version string) (string, string, error) {
	body, err := r.webClient.Get("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml")
	if err != nil {
		return "", "", fmt.Errorf("could not get release yaml: %w", err)
	}

	type YAMLRelease struct {
		Version string `json:"version"`
		URL     struct {
			GZ string `json:"gz"`
		} `json:"url"`
		SHA256 struct {
			GZ string `json:"gz"`
		} `json:"sha256"`
	}

	var releases []YAMLRelease
	err = yaml.Unmarshal(body, &releases)
	if err != nil {
		return "", "", fmt.Errorf("could not unmarshal yaml releases file: %w", err)
	}

	for _, release := range releases {
		if release.Version == version {
			if release.SHA256.GZ != "" && release.URL.GZ != "" {
				return release.URL.GZ, release.SHA256.GZ, nil
			} else {
				break
			}
		}
	}

	return "", "", depErrors.NoSourceCodeError{Version: version}
}

func (r Ruby) getDependencyURLAndSHAFromMirror(version string) (string, string, error) {
	body, err := r.webClient.Get("https://cache.ruby-lang.org/pub/ruby/index.txt")
	if err != nil {
		return "", "", fmt.Errorf("could not get release index: %w", err)
	}

	for _, line := range strings.Split(string(body), "\n") {
		if !strings.HasPrefix(line, "ruby") {
			continue
		}

		versionInfo := strings.Fields(line)
		rubyVersion := strings.TrimPrefix(versionInfo[0], "ruby-")
		if (rubyVersion == version || rubyVersion == version+"-0" || rubyVersion == version+"-p0") &&
			strings.HasSuffix(versionInfo[1], "tar.gz") {
			return versionInfo[1], versionInfo[3], nil
		}
	}

	return "", "", fmt.Errorf("could not find URL and SHA256 for version %s", version)
}
