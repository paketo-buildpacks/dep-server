package dependency

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type DotnetRuntime struct {
	checksummer      Checksummer
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type dotnetRuntimeType struct{}

func (d DotnetRuntime) GetAllVersionRefs() ([]string, error) {
	return dotnet{
		dotnetType:  dotnetRuntimeType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetAllVersionRefs()
}

func (d DotnetRuntime) GetDependencyVersion(version string) (DepVersion, error) {
	return dotnet{
		dotnetType:       dotnetRuntimeType{},
		checksummer:      d.checksummer,
		webClient:        d.webClient,
		licenseRetriever: d.licenseRetriever,
		purlGenerator:    d.purlGenerator,
		name:             "dotnet-runtime",
	}.GetDependencyVersion(version)
}

func (d DotnetRuntime) GetReleaseDate(version string) (*time.Time, error) {
	return dotnet{
		dotnetType:  dotnetRuntimeType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetReleaseDate(version)
}

func (d dotnetRuntimeType) getChannelVersion(version string) string {
	return strings.Join(strings.Split(version, ".")[0:2], ".")
}

func (d dotnetRuntimeType) getReleaseDate(channel DotnetChannel, version string) (*time.Time, error) {
	for _, release := range channel.Releases {
		if release.Runtime.Version == version {
			releaseDate, err := time.Parse("2006-01-02", release.ReleaseDate)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}
			return &releaseDate, nil
		}
	}
	return nil, nil
}

func (d dotnetRuntimeType) getReleaseFiles(channel DotnetChannel, version string) []DotnetChannelReleaseFile {
	for _, release := range channel.Releases {
		if release.Runtime.Version == version {
			return release.Runtime.Files
		}
	}
	return nil
}

func (d dotnetRuntimeType) getReleaseVersions(release DotnetChannelRelease) []string {
	return []string{release.Runtime.Version}
}

func (d dotnetRuntimeType) getCPE(version string) (string, error) {
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

func (d dotnetRuntimeType) versionShouldBeIgnored(version string) bool {
	return false
}
