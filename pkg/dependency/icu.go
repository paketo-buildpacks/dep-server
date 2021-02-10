package dependency

import (
	"encoding/json"
	"errors"
	"fmt"
	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ICU struct {
	checksummer  Checksummer
	fileSystem   FileSystem
	githubClient GithubClient
	webClient    WebClient
}

func (i ICU) GetAllVersionRefs() ([]string, error) {
	releases, err := i.githubClient.GetReleaseTags("unicode-org", "icu")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	for j := 0; j < len(releases); j++ {
		releases[j].TagName = tagToVersion(releases[j].TagName)
	}

	sort.Slice(releases, func(j, k int) bool {
		if releases[j].CreatedDate != releases[k].CreatedDate {
			return releases[j].CreatedDate > releases[k].CreatedDate
		}
		return releases[j].TagName > releases[k].TagName
	})

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.TagName)
	}

	return versions, nil
}

func (i ICU) GetDependencyVersion(version string) (DepVersion, error) {
	releases, err := i.githubClient.GetReleaseTags("unicode-org", "icu")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if tagToVersion(release.TagName) == version {
			depVersion, err := i.createDependencyVersion(version, release)
			if err != nil {
				return DepVersion{}, fmt.Errorf("could not create ICU version: %w", err)
			}
			return depVersion, nil
		}
	}
	return DepVersion{}, fmt.Errorf("could not find ICU version %s", version)
}

func (i ICU) GetReleaseDate(version string) (time.Time, error) {
	releases, err := i.githubClient.GetReleaseTags("unicode-org", "icu")
	if err != nil {
		return time.Time{}, fmt.Errorf("could not get releases: %w", err)
	}

	for _, release := range releases {
		if tagToVersion(release.TagName) == version {
			releaseDate, err := time.Parse(time.RFC3339, release.CreatedDate)
			if err != nil {
				return time.Time{}, fmt.Errorf("could not parse release date: %w", err)
			}
			return releaseDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not find ICU version %s", version)
}

func (i ICU) createDependencyVersion(version string, release internal.GithubRelease) (DepVersion, error) {
	pgpKeysBlock, err := i.webClient.Get("https://raw.githubusercontent.com/unicode-org/icu/master/KEYS")
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get ICU GPG key: %w", err)
	}
	pgpKeys := i.checksummer.SplitPGPKeys(string(pgpKeysBlock))

	icuVersion := versionToICUVersion(version)
	assetName := fmt.Sprintf("icu4c-%s-src.tgz", icuVersion)
	assetDir, err := ioutil.TempDir("", "icu")
	if err != nil {
		return DepVersion{}, fmt.Errorf("failed to create temp directory: %w", err)
	}
	releaseAssetPath := filepath.Join(assetDir, assetName)

	tag := versionToTag(version)
	assetUrl, err := i.githubClient.DownloadReleaseAsset("unicode-org", "icu", tag, assetName, releaseAssetPath)
	if err != nil {
		if errors.Is(err, internal_errors.AssetNotFound{AssetName: assetName}) {
			return DepVersion{}, depErrors.NoSourceCodeError{Version: version}
		}
		return DepVersion{}, fmt.Errorf("could not download asset url: %w", err)
	}

	assetContent, err := i.webClient.Get(assetUrl)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get asset content from asset url: %w", err)
	}

	asset := Asset{}
	err = json.Unmarshal(assetContent, &asset)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not unmarshal asset url content: %w", err)
	}

	assetName = fmt.Sprintf("icu4c-%s-src.tgz.asc", icuVersion)
	releaseAssetSignature, err := i.githubClient.GetReleaseAsset("unicode-org", "icu", tag, assetName)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get release artifact signature: %w", err)
	}

	err = i.checksummer.VerifyASC(string(releaseAssetSignature), releaseAssetPath, pgpKeys...)
	if err != nil {
		return DepVersion{}, fmt.Errorf("release artifact signature verification failed: %w", err)
	}

	dependencySHA, err := i.checksummer.GetSHA256(releaseAssetPath)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get SHA256: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             asset.BrowserDownloadUrl,
		SHA:             dependencySHA,
		ReleaseDate:     release.CreatedDate,
		DeprecationDate: "",
	}, nil
}

func tagToVersion(tagName string) string {
	version := strings.TrimPrefix(tagName, "release-")
	version = strings.ReplaceAll(version, "-", ".")
	return version
}

func versionToICUVersion(version string) string {
	tagName := strings.ReplaceAll(version, ".", "_")
	return tagName
}

func versionToTag(version string) string {
	tag := fmt.Sprintf("release-%s", strings.ReplaceAll(version, ".", "-"))
	return tag
}
