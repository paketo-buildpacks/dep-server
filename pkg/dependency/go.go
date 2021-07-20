package dependency

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
)

type Go struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type GoReleaseWithFiles struct {
	Version string   `json:"version"`
	Files   []GoFile `json:"files"`
}

type GoFile struct {
	SHA256 string `json:"sha256"`
	Kind   string `json:"kind"`
}

type GoRelease struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
}

func (g Go) GetAllVersionRefs() ([]string, error) {
	goReleases, err := g.getGoReleases()
	if err != nil {
		return nil, err
	}

	sort.SliceStable(goReleases, func(i, j int) bool {
		return goReleases[i].ReleaseDate > goReleases[j].ReleaseDate
	})

	var versions []string
	for _, release := range goReleases {
		versions = append(versions, release.Version)
	}

	return versions, nil
}

func (g Go) GetDependencyVersion(version string) (DepVersion, error) {
	goReleasesWithFiles, err := g.getGoReleasesWithFiles()
	if err != nil {
		return DepVersion{}, err
	}

	releaseDate, err := g.GetReleaseDate(version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not find tag for go version %s: %w", version, err)
	}

	sha, err := g.getDependencySHA(version, goReleasesWithFiles)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get dependency SHA256: %w", err)
	}

	depURL := g.dependencyURL(version)

	licenses, err := g.licenseRetriever.LookupLicenses("go", depURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             depURL,
		SHA256:          sha,
		ReleaseDate:     releaseDate,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf("cpe:2.3:a:golang:go:%s:*:*:*:*:*:*:*", strings.TrimPrefix(version, "go")),
		PURL:            g.purlGenerator.Generate("go", version, sha, depURL),
		Licenses:        licenses,
	}, nil
}

func (g Go) GetReleaseDate(version string) (*time.Time, error) {
	body, err := g.webClient.Get("https://golang.org/doc/devel/release.html")
	if err != nil {
		return nil, fmt.Errorf("could not hit golang.org: %w", err)
	}

	re := regexp.MustCompile(fmt.Sprintf(`%s\s*\(released\s*(.*?)\)`, version))
	match := re.FindStringSubmatch(string(body))

	if len(match) < 2 {
		return nil, fmt.Errorf("could not find release date")
	}

	releaseDate, err := time.Parse("2006-01-02", match[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing release date: %w", err)
	}

	return &releaseDate, nil
}

func (g Go) getDependencySHA(version string, releases []GoReleaseWithFiles) (string, error) {
	sha := ""
	foundSHA := false
	for _, release := range releases {
		if release.Version == version {
			for _, file := range release.Files {
				if file.Kind == "source" {
					sha = file.SHA256
					foundSHA = true
				}
			}
		}
	}

	if !foundSHA {
		return "", fmt.Errorf("could not find SHA256 for %s: %w", version, errors.NoSourceCodeError{Version: version})
	}

	if sha == "" {
		return g.calculateDependencySHA(version)
	}

	return sha, nil
}

func (g Go) calculateDependencySHA(version string) (string, error) {
	tempDir, err := ioutil.TempDir("", "httpd")
	if err != nil {
		return "", fmt.Errorf("could not make temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	url := g.dependencyURL(version)
	dependencyPath := filepath.Join(tempDir, filepath.Base(url))

	err = g.webClient.Download(url, dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	sha256, err := g.checksummer.GetSHA256(dependencyPath)
	if err != nil {
		return "", fmt.Errorf("could not get sha256: %w", err)
	}

	return sha256, nil
}

func (g Go) getGoReleases() ([]GoRelease, error) {
	body, err := g.webClient.Get("https://golang.org/doc/devel/release.html")
	if err != nil {
		return nil, fmt.Errorf("could not hit golang.org: %w", err)
	}

	re := regexp.MustCompile(`>\s*?(go[0-9.]*?)\s*\(released\s*(.*?)\)`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	var goReleases []GoRelease
	for _, match := range matches {
		goReleases = append(goReleases, GoRelease{
			Version:     match[1],
			ReleaseDate: match[2],
		})
	}

	return g.removeReleasesWithoutFiles(goReleases)
}

func (g Go) removeReleasesWithoutFiles(allReleases []GoRelease) ([]GoRelease, error) {
	releasesWithFiles, err := g.getGoReleasesWithFiles()
	if err != nil {
		return nil, fmt.Errorf("could not get go releases with files: %w", err)
	}

	releaseMap := make(map[string]bool)
	for _, release := range releasesWithFiles {
		releaseMap[release.Version] = true
	}

	var prunedReleases []GoRelease
	for _, release := range allReleases {
		if releaseMap[release.Version] == true {
			prunedReleases = append(prunedReleases, release)
		}
	}

	return prunedReleases, nil
}

func (g Go) getGoReleasesWithFiles() ([]GoReleaseWithFiles, error) {
	body, err := g.webClient.Get("https://golang.org/dl/?mode=json&include=all")
	if err != nil {
		return nil, fmt.Errorf("could not hit golang.org: %w", err)
	}

	var goReleasesWithFiles []GoReleaseWithFiles
	err = json.Unmarshal(body, &goReleasesWithFiles)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response: %w\n%s", err, body)
	}

	return goReleasesWithFiles, nil
}

func (g Go) dependencyURL(version string) string {
	return fmt.Sprintf("https://dl.google.com/go/%s.src.tar.gz", version)
}
