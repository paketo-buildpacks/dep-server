package dependency

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	depErrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal/internal_errors"
)

type ICU struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	githubClient     GithubClient
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

func (i ICU) GetAllVersionRefs() ([]string, error) {
	releases, err := i.getAllVersions()
	if err != nil {
		return nil, err
	}

	for j := 0; j < len(releases); j++ {
		releases[j].TagName = tagToVersion(releases[j].TagName)
	}

	sort.Slice(releases, func(j, k int) bool {
		if releases[j].CreatedDate != releases[k].CreatedDate {
			return releases[j].CreatedDate.After(releases[k].CreatedDate)
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
	releases, err := i.getAllVersions()
	if err != nil {
		return DepVersion{}, err
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

func (i ICU) GetReleaseDate(version string) (*time.Time, error) {
	releases, err := i.getAllVersions()
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		if tagToVersion(release.TagName) == version {
			return &release.CreatedDate, nil
		}
	}

	return nil, fmt.Errorf("could not find ICU version %s", version)
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
		if errors.Is(err, internal_errors.AssetNotFound{AssetName: assetName}) {
			return DepVersion{}, depErrors.NoSourceCodeError{Version: version}
		}
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

	licenses, err := i.licenseRetriever.LookupLicenses("icu", asset.BrowserDownloadUrl)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             asset.BrowserDownloadUrl,
		SHA256:          dependencySHA,
		ReleaseDate:     &release.CreatedDate,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf(`cpe:2.3:a:icu-project:international_components_for_unicode:%s:*:*:*:*:c\/c\+\+:*:*`, version),
		PURL:            i.purlGenerator.Generate("icu", version, dependencySHA, asset.BrowserDownloadUrl),
		Licenses:        licenses,
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

func (i ICU) getAllVersions() ([]internal.GithubRelease, error) {
	releases, err := i.githubClient.GetReleaseTags("unicode-org", "icu")
	if err != nil {
		return nil, fmt.Errorf("could not get releases: %w", err)
	}

	tagsToIgnore := map[string]bool{
		"release-59-1": true,
		"release-58-2": true,
		"release-57-1": true,
		"release-56-1": true,
	}

	var prunedReleases []internal.GithubRelease
	for _, release := range releases {
		if !tagsToIgnore[release.TagName] {
			prunedReleases = append(prunedReleases, release)
		}
	}
	return prunedReleases, nil
}
