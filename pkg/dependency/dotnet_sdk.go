package dependency

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type DotnetSDK struct {
	checksummer      Checksummer
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type dotnetSDKType struct{}

func (d DotnetSDK) GetAllVersionRefs() ([]string, error) {
	return dotnet{
		dotnetType:  dotnetSDKType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetAllVersionRefs()
}

func (d DotnetSDK) GetDependencyVersion(version string) (DepVersion, error) {
	return dotnet{
		dotnetType:       dotnetSDKType{},
		checksummer:      d.checksummer,
		webClient:        d.webClient,
		licenseRetriever: d.licenseRetriever,
		purlGenerator:    d.purlGenerator,
		name:             "dotnet-sdk",
	}.GetDependencyVersion(version)
}

func (d DotnetSDK) GetReleaseDate(version string) (*time.Time, error) {
	return dotnet{
		dotnetType:  dotnetSDKType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetReleaseDate(version)
}

func (d dotnetSDKType) getReleaseFiles(channel DotnetChannel, version string) []DotnetChannelReleaseFile {
	for _, release := range channel.Releases {
		if release.SDK.Version == version {
			return release.SDK.Files
		}

		for _, sdk := range release.SDKs {
			if sdk.Version == version {
				return sdk.Files
			}
		}
	}
	return nil
}

func (d dotnetSDKType) getReleaseDate(channel DotnetChannel, version string) (*time.Time, error) {
	for _, release := range channel.Releases {
		if release.SDK.Version == version {
			releaseDate, err := time.Parse("2006-01-02", release.ReleaseDate)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}
			return &releaseDate, nil
		}

		for _, sdk := range release.SDKs {
			if sdk.Version == version {
				releaseDate, err := time.Parse("2006-01-02", release.ReleaseDate)
				if err != nil {
					return nil, fmt.Errorf("could not parse release date: %w", err)
				}
				return &releaseDate, nil
			}
		}
	}
	return nil, nil
}

func (d dotnetSDKType) getReleaseVersions(release DotnetChannelRelease) []string {
	versions := []string{release.SDK.Version}
	for _, sdk := range release.SDKs {
		versions = append(versions, sdk.Version)
	}
	return versions
}

func (d dotnetSDKType) getChannelVersion(version string) string {
	versionsInWrongChannel := map[string]string{
		"2.1.201":                 "2.0",
		"2.1.200":                 "2.0",
		"2.1.105":                 "2.0",
		"2.1.104":                 "2.0",
		"2.1.103":                 "2.0",
		"2.1.102":                 "2.0",
		"2.1.101":                 "2.0",
		"2.1.100":                 "2.0",
		"2.1.4":                   "2.0",
		"2.1.3":                   "2.0",
		"2.1.2":                   "2.0",
		"1.0.0-preview2.1-003177": "1.1",
	}

	if channelVersion, ok := versionsInWrongChannel[version]; ok {
		return channelVersion
	}

	return strings.Join(strings.Split(version, ".")[0:2], ".")
}

func (d dotnetSDKType) getCPE(version string) (string, error) {
	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to parse semverL %w", err)
	}

	productName := ".net"
	if parsedVersion.LessThan(semver.MustParse("5.0.0-0")) { // use 5.0.0-0 to ensure 5.0.0 previews/RCs use the new `.net` product name
		productName = ".net_core"
	}
	return fmt.Sprintf("cpe:2.3:a:microsoft:%s:%s:*:*:*:*:*:*:*", productName, version), nil
}

func (d dotnetSDKType) versionShouldBeIgnored(version string) bool {
	versionsWithWrongHash := map[string]bool{
		"2.1.202": true,
		"1.1.5":   true,
	}

	_, shouldBeIgnored := versionsWithWrongHash[version]
	return shouldBeIgnored
}
