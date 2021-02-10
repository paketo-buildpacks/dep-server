package dependency

import "strings"

type DotnetSDK struct {
	checksummer Checksummer
	webClient   WebClient
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
		dotnetType:  dotnetSDKType{},
		checksummer: d.checksummer,
		webClient:   d.webClient,
	}.GetDependencyVersion(version)
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

func (d dotnetSDKType) getReleaseDate(channel DotnetChannel, version string) string {
	for _, release := range channel.Releases {
		if release.SDK.Version == version {
			return release.ReleaseDate
		}

		for _, sdk := range release.SDKs {
			if sdk.Version == version {
				return release.ReleaseDate
			}
		}
	}
	return ""
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

func (d dotnetSDKType) versionShouldBeIgnored(version string) bool {
	versionsWithWrongHash := map[string]bool{
		"2.1.202": true,
		"1.1.5":   true,
	}

	_, shouldBeIgnored := versionsWithWrongHash[version]
	return shouldBeIgnored
}
