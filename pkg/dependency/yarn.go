package dependency

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"

	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
)

type Yarn struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	githubClient     GithubClient
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

type YarnRelease struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}
type Asset struct {
	BrowserDownloadUrl string `json:"browser_download_url"`
}

func (y Yarn) GetAllVersionRefs() ([]string, error) {
	releases, err := y.githubClient.GetReleaseTags("yarnpkg", "yarn")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	var versions []string
	for _, release := range releases {
		versionTagName := strings.TrimPrefix(release.TagName, "v")
		version, err := semver.NewVersion(versionTagName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version: %w", err)
		}
		/** Versions less than 0.7.0 does not have source code and the version tag does not contains the "v" at the start*/
		if version.LessThan(semver.MustParse("0.7.0")) {
			continue
		}
		if version.Prerelease() != "" {
			continue
		}
		versions = append(versions, versionTagName)
	}
	return versions, nil
}

func (y Yarn) GetDependencyVersion(version string) (DepVersion, error) {
	releases, err := y.githubClient.GetReleaseTags("yarnpkg", "yarn")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		tagName := "v" + version
		if release.TagName == tagName {
			depVersion, err := y.createDependencyVersion(version, tagName, release)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not create yarn version: %w", err)
			}
			return depVersion, nil
		}
	}
	return DepVersion{}, fmt.Errorf("could not find yarn version %s", version)
}

func (y Yarn) GetReleaseDate(version string) (*time.Time, error) {
	releases, err := y.githubClient.GetReleaseTags("yarnpkg", "yarn")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.TagName == version {
			return &release.PublishedDate, nil
		}
	}

	return nil, fmt.Errorf("could not find release date for version %s", version)
}

func (y Yarn) createDependencyVersion(version, tagName string, release internal.GithubRelease) (DepVersion, error) {
	yarnGPGKey, err := y.webClient.Get("https://dl.yarnpkg.com/debian/pubkey.gpg")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get yarn GPG key: %w", err)
	}

	releaseAssetDir, err := ioutil.TempDir("", "yarn")
	if err != nil {
		return DepVersion{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(releaseAssetDir)
	releaseAssetPath := filepath.Join(releaseAssetDir, fmt.Sprintf("yarn-%s.tar.gz", tagName))

	assetName := fmt.Sprintf("yarn-%s.tar.gz", tagName)
	assetUrl, err := y.githubClient.DownloadReleaseAsset("yarnpkg", "yarn", tagName, assetName, releaseAssetPath)
	if err != nil {
		if errors.Is(err, internal_errors.AssetNotFound{AssetName: assetName}) {
			return DepVersion{}, depErrors.NoSourceCodeError{Version: version}
		}
		return DepVersion{}, fmt.Errorf("could not download asset url: %w", err)
	}

	assetContent, err := y.webClient.Get(assetUrl)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get asset content from asset url: %w", err)
	}

	asset := Asset{}
	err = json.Unmarshal(assetContent, &asset)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not unmarshal asset url content: %w", err)
	}

	assetName = fmt.Sprintf("yarn-%s.tar.gz.asc", tagName)
	releaseAssetSignature, err := y.githubClient.GetReleaseAsset("yarnpkg", "yarn", tagName, assetName)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release artifact signature: %w", err)
	}

	err = y.checksummer.VerifyASC(string(releaseAssetSignature), releaseAssetPath, string(yarnGPGKey))
	if err != nil {
		return DepVersion{}, fmt.Errorf("release artifact signature verification failed: %w", err)
	}

	dependencySHA, err := y.checksummer.GetSHA256(releaseAssetPath)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get SHA256: %w", err)
	}

	licenses, err := y.licenseRetriever.LookupLicenses("yarn", asset.BrowserDownloadUrl)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             asset.BrowserDownloadUrl,
		SHA256:          dependencySHA,
		ReleaseDate:     &release.PublishedDate,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf("cpe:2.3:a:yarnpkg:yarn:%s:*:*:*:*:*:*:*", version),
		PURL:            y.purlGenerator.Generate("yarn", version, dependencySHA, asset.BrowserDownloadUrl),
		Licenses:        licenses,
	}, nil
}
