package dependency

import (
	"fmt"
	"strings"
	"time"
)

type DotnetASPNETCore struct {
	checksummer      Checksummer
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type dotnetASPNETCoreType struct{}

func (d DotnetASPNETCore) GetAllVersionRefs() ([]string, error) {
	return dotnet{
		dotnetType:  dotnetASPNETCoreType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetAllVersionRefs()
}

func (d DotnetASPNETCore) GetDependencyVersion(version string) (DepVersion, error) {
	return dotnet{
		dotnetType:       dotnetASPNETCoreType{},
		checksummer:      d.checksummer,
		webClient:        d.webClient,
		licenseRetriever: d.licenseRetriever,
		purlGenerator:    d.purlGenerator,
		name:             "dotnet-aspnetcore",
	}.GetDependencyVersion(version)
}

func (d DotnetASPNETCore) GetReleaseDate(version string) (*time.Time, error) {
	return dotnet{
		dotnetType:  dotnetASPNETCoreType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetReleaseDate(version)
}

func (d dotnetASPNETCoreType) getChannelVersion(version string) string {
	return strings.Join(strings.Split(version, ".")[0:2], ".")
}

func (d dotnetASPNETCoreType) getReleaseDate(channel DotnetChannel, version string) (*time.Time, error) {
	for _, release := range channel.Releases {
		if release.ASPNETCoreRuntime.Version == version {
			releaseDate, err := time.Parse("2006-01-02", release.ReleaseDate)
			if err != nil {
				return nil, fmt.Errorf("could not parse release date: %w", err)
			}
			return &releaseDate, nil
		}
	}
	return nil, nil
}

func (d dotnetASPNETCoreType) getReleaseFiles(channel DotnetChannel, version string) []DotnetChannelReleaseFile {
	for _, release := range channel.Releases {
		if release.ASPNETCoreRuntime.Version == version {
			return release.ASPNETCoreRuntime.Files
		}
	}
	return nil
}

func (d dotnetASPNETCoreType) getReleaseVersions(release DotnetChannelRelease) []string {
	return []string{release.ASPNETCoreRuntime.Version}
}

func (d dotnetASPNETCoreType) getCPE(version string) (string, error) {
	majorMinorVersion := strings.Join(strings.Split(version, ".")[0:2], ".")
	return fmt.Sprintf("cpe:2.3:a:microsoft:asp.net_core:%s:*:*:*:*:*:*:*", majorMinorVersion), nil
}

func (d dotnetASPNETCoreType) versionShouldBeIgnored(version string) bool {
	return false
}
