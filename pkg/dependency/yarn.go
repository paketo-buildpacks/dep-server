package dependency

import (
	"encoding/json"
	"errors"
	"fmt"
	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Yarn struct {
	checksummer  Checksummer
	fileSystem   FileSystem
	githubClient GithubClient
	webClient    WebClient
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
		versions = append(versions, release.TagName)
	}
	return versions, nil
}

func (y Yarn) GetDependencyVersion(version string) (DepVersion, error) {
	releases, err := y.githubClient.GetReleaseTags("yarnpkg", "yarn")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if release.TagName == version {
			depVersion, err := y.createDependencyVersion(version, release)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not create yarn version: %w", err)
			}
			return depVersion, nil
		}
	}
	return DepVersion{}, fmt.Errorf("could not find yarn version %s", version)
}

func (y Yarn) createDependencyVersion(version string, release internal.GithubRelease) (DepVersion, error) {
	yarnGPGKey, err := y.webClient.Get("https://dl.yarnpkg.com/debian/pubkey.gpg")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get yarn GPG key: %w", err)
	}

	releaseAssetDir, err := ioutil.TempDir("", "yarn")
	if err != nil {
		return DepVersion{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(releaseAssetDir)
	releaseAssetPath := filepath.Join(releaseAssetDir, fmt.Sprintf("yarn-%s.tar.gz", version))

	assetName := fmt.Sprintf("yarn-%s.tar.gz", version)
	assetUrl, err := y.githubClient.DownloadReleaseAsset("yarnpkg", "yarn", version, assetName, releaseAssetPath)
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

	assetName = fmt.Sprintf("yarn-%s.tar.gz.asc", version)
	releaseAssetSignature, err := y.githubClient.GetReleaseAsset("yarnpkg", "yarn", version, assetName)
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

	return DepVersion{
		Version:         version,
		URI:             asset.BrowserDownloadUrl,
		SHA:             dependencySHA,
		ReleaseDate:     release.PublishedDate,
		DeprecationDate: "",
	}, nil
}
