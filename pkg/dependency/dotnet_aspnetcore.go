package dependency

import (
	"strings"
	"time"
)

type DotnetASPNETCore struct {
	checksummer Checksummer
	webClient   WebClient
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
		dotnetType:  dotnetASPNETCoreType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetDependencyVersion(version)
}

func (d DotnetASPNETCore) GetReleaseDate(version string) (time.Time, error) {
	return dotnet{
		dotnetType:  dotnetASPNETCoreType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetReleaseDate(version)
}

func (d dotnetASPNETCoreType) getChannelVersion(version string) string {
	return strings.Join(strings.Split(version, ".")[0:2], ".")
}

func (d dotnetASPNETCoreType) getReleaseDate(channel DotnetChannel, version string) string {
	for _, release := range channel.Releases {
		if release.ASPNETCoreRuntime.Version == version {
			return release.ReleaseDate
		}
	}
	return ""
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

func (d dotnetASPNETCoreType) versionShouldBeIgnored(version string) bool {
	return false
}
