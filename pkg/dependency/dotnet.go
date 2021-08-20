package dependency

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
)

const (
	DotnetReleaseIndexURL = "https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/releases-index.json"
	DotnetChannelURL      = "https://dotnetcli.blob.core.windows.net/dotnet/release-metadata/%s/releases.json"
	DotnetLinuxFileRID    = "linux-x64"
	DotnetUbuntuFileRID   = "ubuntu-x64"
)

type DotnetChannel struct {
	EOLDate  string                 `json:"eol-date"`
	Releases []DotnetChannelRelease `json:"releases"`
}

type DotnetChannelRelease struct {
	ReleaseDate       string `json:"release-date"`
	ASPNETCoreRuntime struct {
		Version string                     `json:"version"`
		Files   []DotnetChannelReleaseFile `json:"files"`
	} `json:"aspnetcore-runtime"`
	Runtime struct {
		Version string                     `json:"version"`
		Files   []DotnetChannelReleaseFile `json:"files"`
	} `json:"runtime"`
	SDK struct {
		Version string                     `json:"version"`
		Files   []DotnetChannelReleaseFile `json:"files"`
	} `json:"sdk"`
	SDKs []struct {
		Version string                     `json:"version"`
		Files   []DotnetChannelReleaseFile `json:"files"`
	} `json:"sdks"`
}

type DotnetVersion struct {
	Version     string
	ReleaseDate string
}

type DotnetChannelReleaseFile struct {
	Name string `json:"name"`
	RID  string `json:"rid"`
	URL  string `json:"url"`
	Hash string `json:"hash"`
}

type dotnetType interface {
	getChannelVersion(version string) string
	getReleaseDate(channel DotnetChannel, version string) (*time.Time, error)
	getReleaseFiles(channel DotnetChannel, version string) []DotnetChannelReleaseFile
	getReleaseVersions(release DotnetChannelRelease) []string
	versionShouldBeIgnored(version string) bool
	getCPE(version string) (string, error)
}

type dotnet struct {
	dotnetType       dotnetType
	checksummer      Checksummer
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
	name             string
}

func (d dotnet) GetAllVersionRefs() ([]string, error) {
	channelVersions, err := d.getAllChannelVersions()
	if err != nil {
		return nil, fmt.Errorf("could not get channel sortedVersions: %w", err)
	}

	var versions []DotnetVersion
	for _, channelVersion := range channelVersions {
		versionsForChannel, err := d.getVersionsForChannel(channelVersion)
		if err != nil {
			return nil, fmt.Errorf("could not get sortedVersions for channel %s: %w", channelVersion, err)
		}

		versions = append(versions, versionsForChannel...)
	}

	sortedVersions, err := d.sortVersions(versions)
	if err != nil {
		return nil, fmt.Errorf("could not sort versions: %w", err)
	}

	return sortedVersions, nil
}

func (d dotnet) GetDependencyVersion(version string) (DepVersion, error) {
	channel, err := d.getChannel(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get channel: %w", err)
	}

	releaseFile, err := d.getReleaseFile(channel, version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release file: %w", err)
	}

	sha256, err := d.getReleaseFileSHA(releaseFile)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get sha: %w", err)
	}

	releaseDate, err := d.dotnetType.getReleaseDate(channel, version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("error getting release date: %w", err)
	}

	cpe, err := d.dotnetType.getCPE(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get cpe: %w", err)
	}

	licenses, err := d.licenseRetriever.LookupLicenses(d.name, releaseFile.URL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get licenses: %w", err)
	}

	depVersion := DepVersion{
		Version:     version,
		URI:         releaseFile.URL,
		SHA256:      sha256,
		ReleaseDate: releaseDate,
		CPE:         cpe,
		PURL:        d.purlGenerator.Generate(d.name, version, sha256, releaseFile.URL),
		Licenses:    licenses,
	}
	if channel.EOLDate != "" {
		deprecationDate, err := time.Parse("2006-01-02", channel.EOLDate)
		if err != nil {
			return DepVersion{}, fmt.Errorf("could not parse EOL date: %w", err)
		}
		depVersion.DeprecationDate = &deprecationDate
	}
	return depVersion, nil
}

func (d dotnet) GetReleaseDate(version string) (*time.Time, error) {
	channel, err := d.getChannel(version)
	if err != nil {
		return nil, fmt.Errorf("could not get channel: %w", err)
	}

	return d.dotnetType.getReleaseDate(channel, version)
}

func (d dotnet) getAllChannelVersions() ([]string, error) {
	body, err := d.webClient.Get(DotnetReleaseIndexURL)
	if err != nil {
		return nil, fmt.Errorf("could not get releases index body: %w", err)
	}

	var releasesIndex struct {
		Channels []struct {
			ChannelVersion string `json:"channel-version"`
		} `json:"releases-index"`
	}
	err = json.Unmarshal(body, &releasesIndex)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal releases index: %w", err)
	}

	var channelVersions []string
	for _, channel := range releasesIndex.Channels {
		channelVersions = append(channelVersions, channel.ChannelVersion)
	}
	return channelVersions, nil
}

func (d dotnet) getVersionsForChannel(version string) ([]DotnetVersion, error) {
	channel, err := d.getChannel(version)
	if err != nil {
		return nil, fmt.Errorf("could not get channel: %w", err)
	}

	uniqueVersions := make(map[string]bool)
	var versions []DotnetVersion
	for i := len(channel.Releases) - 1; i >= 0; i-- {
		release := channel.Releases[i]

		versionsForRelease := d.dotnetType.getReleaseVersions(release)
		for _, version := range versionsForRelease {
			if version == "" || d.dotnetType.versionShouldBeIgnored(version) {
				continue
			}
			parsedVersion, err := semver.NewVersion(version)
			if err != nil {
				return nil, fmt.Errorf("failed to parse version: %w", err)
			}
			if parsedVersion.Prerelease() != "" {
				continue
			}

			if _, exists := uniqueVersions[version]; !exists {
				versions = append(versions, DotnetVersion{Version: version, ReleaseDate: release.ReleaseDate})
				uniqueVersions[version] = true
			}
		}
	}

	return versions, nil
}

func (d dotnet) getChannel(version string) (DotnetChannel, error) {
	channelVersion := d.dotnetType.getChannelVersion(version)

	body, err := d.webClient.Get(fmt.Sprintf(DotnetChannelURL, channelVersion))
	if err != nil {
		return DotnetChannel{}, fmt.Errorf("could not get channel body: %w", err)
	}

	var channel DotnetChannel
	err = json.Unmarshal(body, &channel)
	if err != nil {
		return DotnetChannel{}, fmt.Errorf("could not unmarshal channel: %w", err)
	}

	return channel, nil
}

func (d dotnet) sortVersions(versions []DotnetVersion) ([]string, error) {
	var sortErr error
	sort.Slice(versions, func(i, j int) bool {
		if versions[i].ReleaseDate != versions[j].ReleaseDate {
			return versions[i].ReleaseDate < versions[j].ReleaseDate
		}

		semver1, err := semver.NewVersion(versions[i].Version)
		if err != nil {
			sortErr = fmt.Errorf("could not parse '%s' as semver", versions[i].Version)
			return false
		}
		semver2, err := semver.NewVersion(versions[j].Version)
		if err != nil {
			sortErr = fmt.Errorf("could not parse '%s' as semver", versions[j].Version)
			return false
		}
		return semver1.LessThan(semver2)
	})
	if sortErr != nil {
		return nil, sortErr
	}

	uniqueVersions := make(map[string]bool)
	var sortedVersions []string
	for _, version := range versions {
		if _, exists := uniqueVersions[version.Version]; !exists {
			sortedVersions = append([]string{version.Version}, sortedVersions...)
			uniqueVersions[version.Version] = true
		}
	}

	return sortedVersions, nil
}

func (d dotnet) getReleaseFile(channel DotnetChannel, version string) (DotnetChannelReleaseFile, error) {
	files := d.dotnetType.getReleaseFiles(channel, version)
	for _, file := range files {
		if file.RID == DotnetLinuxFileRID {
			return file, nil
		}
	}
	for _, file := range files {
		if file.RID == DotnetUbuntuFileRID {
			return file, nil
		}
	}

	return DotnetChannelReleaseFile{}, errors.NoSourceCodeError{Version: version}
}

func (d dotnet) getReleaseFileSHA(file DotnetChannelReleaseFile) (string, error) {
	if len(file.Hash) == 64 {
		return strings.ToLower(file.Hash), nil
	}

	tempDir, err := ioutil.TempDir("", "nginx")
	if err != nil {
		return "", fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	dependencyOutputPath := filepath.Join(tempDir, file.Name)
	err = d.webClient.Download(file.URL, dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	if file.Hash != "" {
		err = d.checksummer.VerifySHA512(dependencyOutputPath, strings.ToLower(file.Hash))
		if err != nil {
			return "", fmt.Errorf("dependency signature verification failed: %w", err)
		}
	}

	sha256, err := d.checksummer.GetSHA256(dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not get SHA256: %w", err)
	}

	return sha256, nil
}
