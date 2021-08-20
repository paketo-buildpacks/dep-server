package dependency

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Node struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type NodeRelease struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}

type ReleaseSchedule map[string]struct {
	End string `json:"end"`
}

func (n Node) GetAllVersionRefs() ([]string, error) {
	nodeReleases, err := n.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get node releases: %w", err)
	}

	sort.SliceStable(nodeReleases, func(i, j int) bool {
		return nodeReleases[i].Date > nodeReleases[j].Date
	})

	var versions []string
	for _, release := range nodeReleases {
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (n Node) GetDependencyVersion(version string) (DepVersion, error) {
	nodeReleases, err := n.getAllReleases()
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	releaseSchedule, err := n.getReleaseSchedule()
	if err != nil {
		return DepVersion{}, err
	}

	for _, release := range nodeReleases {
		if release.Version == version {
			depVersion, err := n.createDepVersion(release, releaseSchedule)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not create dep version: %w", err)
			}
			return depVersion, nil
		}
	}

	return DepVersion{}, fmt.Errorf("could not find version %s", version)
}

func (n Node) GetReleaseDate(version string) (*time.Time, error) {
	nodeReleases, err := n.getAllReleases()
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range nodeReleases {
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

func (n Node) getAllReleases() ([]NodeRelease, error) {
	body, err := n.webClient.Get("https://nodejs.org/dist/index.json")
	if err != nil {
		return nil, fmt.Errorf("could not get release index: %w", err)
	}

	var nodeReleases []NodeRelease
	err = json.Unmarshal(body, &nodeReleases)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w\n%s", err, body)
	}

	return nodeReleases, nil
}

func (n Node) createDepVersion(release NodeRelease, releaseSchedule ReleaseSchedule) (DepVersion, error) {
	deprecationDate := n.getDeprecationDate(release.Version, releaseSchedule)
	sha, err := n.getDependencySHA(release.Version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get dependency SHA256: %w", err)
	}

	releaseDate, err := time.Parse("2006-01-02", release.Date)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not parse release date: %w", err)
	}
	depURL := n.dependencyURL(release.Version)
	licenses, err := n.licenseRetriever.LookupLicenses("node", depURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         release.Version,
		URI:             depURL,
		SHA256:          sha,
		ReleaseDate:     &releaseDate,
		DeprecationDate: deprecationDate,
		CPE:             fmt.Sprintf("cpe:2.3:a:nodejs:node.js:%s:*:*:*:*:*:*:*", strings.TrimPrefix(release.Version, "v")),
		PURL:            n.purlGenerator.Generate("node", release.Version, sha, depURL),
		Licenses:        licenses,
	}, nil
}

func (n Node) getReleaseSchedule() (ReleaseSchedule, error) {
	body, err := n.webClient.Get("https://raw.githubusercontent.com/nodejs/Release/master/schedule.json")
	if err != nil {
		return ReleaseSchedule{}, fmt.Errorf("could not get release schedule: %w", err)
	}

	var releaseSchedule map[string]struct {
		End string `json:"end"`
	}
	err = json.Unmarshal(body, &releaseSchedule)
	if err != nil {
		return ReleaseSchedule{}, fmt.Errorf("could not unmarshal release schedule: %w\n%s", err, body)
	}

	return releaseSchedule, nil
}

func (n Node) getDeprecationDate(version string, releaseSchedule ReleaseSchedule) *time.Time {
	versionIndex := strings.Split(version, ".")[0]
	if versionIndex == "v0" {
		versionIndex = strings.Join(strings.Split(version, ".")[0:2], ".")
	}
	release, ok := releaseSchedule[versionIndex]
	if !ok {
		return nil
	}

	deprecationDate, err := time.Parse("2006-01-02", release.End)
	if err != nil {
		return nil
	}

	return &deprecationDate
}

func (n Node) getDependencySHA(version string) (string, error) {
	body, err := n.webClient.Get(n.shaFileURL(version))
	if err != nil {
		return "", fmt.Errorf("could not get SHA256 file: %w", err)
	}

	var dependencySHA string
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasSuffix(line, fmt.Sprintf("node-%s.tar.gz", version)) {
			dependencySHA = strings.Fields(line)[0]
		}
	}
	if dependencySHA == "" {
		return "", fmt.Errorf("could not find SHA256 for node-%s.tar.gz", version)
	}
	return dependencySHA, nil
}

func (n Node) dependencyURL(version string) string {
	return fmt.Sprintf("https://nodejs.org/dist/%s/node-%s.tar.gz", version, version)
}

func (n Node) shaFileURL(version string) string {
	return fmt.Sprintf("https://nodejs.org/dist/%s/SHASUMS256.txt", version)
}
