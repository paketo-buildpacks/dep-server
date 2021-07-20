package dependency

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type Rust struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	githubClient     GithubClient
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

func (r Rust) GetAllVersionRefs() ([]string, error) {
	tags, err := r.githubClient.GetTags("rust-lang", "rust")
	if err != nil {
		return nil, fmt.Errorf("could not get tags: %w", err)
	}

	var versions []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "release-") {
			continue
		}
		version, err := semver.NewVersion(tag)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version %s: %w", tag, err)
		}
		if version.Prerelease() != "" || version.Major() == 0 {
			continue
		}
		versions = append(versions, tag)
	}
	return versions, nil
}

func (r Rust) GetDependencyVersion(version string) (DepVersion, error) {
	releaseDate, err := r.GetReleaseDate(version)
	if err != nil {
		return DepVersion{}, err
	}

	dependencyURL := r.dependencyURL(version)
	sha, err := r.getDependencySHA(dependencyURL, version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get rust sha: %w", err)
	}

	licenses, err := r.licenseRetriever.LookupLicenses("rust", dependencyURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             dependencyURL,
		SHA256:          sha,
		ReleaseDate:     releaseDate,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf("cpe:2.3:a:rust-lang:rust:%s:*:*:*:*:*:*:*", version),
		PURL:            r.purlGenerator.Generate("rust", version, sha, dependencyURL),
		Licenses:        licenses,
	}, nil
}

func (r Rust) GetReleaseDate(version string) (*time.Time, error) {
	tagCommit, err := r.githubClient.GetTagCommit("rust-lang", "rust", version)
	if err != nil {
		return nil, fmt.Errorf("could not get release date: %w", err)
	}

	return &tagCommit.Date, nil
}

func (r Rust) getDependencySHA(dependencyURL, version string) (string, error) {
	rustGPGKey, err := r.webClient.Get("https://static.rust-lang.org/rust-key.gpg.ascii")
	if err != nil {
		return "", fmt.Errorf("could not get rust GPG key: %w", err)
	}

	dependencySignature, err := r.webClient.Get(r.dependencySignatureURL(version))
	if err != nil {
		return "", fmt.Errorf("could not get dependency signature: %w", err)
	}

	dependencyOutputDir, err := ioutil.TempDir("", "rust")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	dependencyOutputPath := filepath.Join(dependencyOutputDir, filepath.Base(dependencyURL))

	err = r.webClient.Download(dependencyURL, dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	err = r.checksummer.VerifyASC(string(dependencySignature), dependencyOutputPath, string(rustGPGKey))
	if err != nil {
		return "", fmt.Errorf("dependency signature verification failed: %w", err)
	}

	dependencySHA, err := r.checksummer.GetSHA256(dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not get SHA256: %w", err)
	}

	return dependencySHA, nil
}

func (r Rust) dependencySignatureURL(version string) string {
	return fmt.Sprintf("https://static.rust-lang.org/dist/rustc-%s-src.tar.gz.asc", version)
}

func (r Rust) dependencyURL(version string) string {
	return fmt.Sprintf("https://static.rust-lang.org/dist/rustc-%s-src.tar.gz", version)
}
