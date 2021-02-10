package dependency

import (
	"strings"
	"time"
)

type DotnetRuntime struct {
	checksummer Checksummer
	webClient   WebClient
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
		dotnetType:  dotnetRuntimeType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetDependencyVersion(version)
}

func (d DotnetRuntime) GetReleaseDate(version string) (time.Time, error) {
	return dotnet{
		dotnetType:  dotnetRuntimeType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetReleaseDate(version)
}

func (d dotnetRuntimeType) getChannelVersion(version string) string {
	return strings.Join(strings.Split(version, ".")[0:2], ".")
}

func (d dotnetRuntimeType) getReleaseDate(channel DotnetChannel, version string) string {
	for _, release := range channel.Releases {
		if release.Runtime.Version == version {
			return release.ReleaseDate
		}
	}
	return ""
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

func (d dotnetRuntimeType) versionShouldBeIgnored(version string) bool {
	return false
}
