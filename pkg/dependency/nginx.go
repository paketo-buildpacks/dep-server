package dependency

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"
)

type Nginx struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	githubClient     GithubClient
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

func (n Nginx) GetAllVersionRefs() ([]string, error) {
	tags, err := n.githubClient.GetTags("nginx", "nginx")
	if err != nil {
		return nil, fmt.Errorf("could not get tags: %w", err)
	}

	var versions []string
	for _, tag := range tags {
		versions = append(versions, strings.TrimPrefix(tag, "release-"))
	}

	return versions, nil
}

func (n Nginx) GetDependencyVersion(version string) (DepVersion, error) {
	tagName := "release-" + version
	tagCommit, err := n.githubClient.GetTagCommit("nginx", "nginx", tagName)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get tags: %w", err)
	}

	dependencyURL := n.dependencyURL(version)
	sha, err := n.getDependencySHA(dependencyURL, version)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get nginx sha: %w", err)
	}

	licenses, err := n.licenseRetriever.LookupLicenses("nginx", dependencyURL)
	if err != nil {
		return DepVersion{}, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	return DepVersion{
		Version:         version,
		URI:             dependencyURL,
		SHA256:          sha,
		ReleaseDate:     &tagCommit.Date,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf("cpe:2.3:a:nginx:nginx:%s:*:*:*:*:*:*:*", version),
		PURL:            n.purlGenerator.Generate("nginx", version, sha, dependencyURL),
		Licenses:        licenses,
	}, nil
}

func (n Nginx) GetReleaseDate(version string) (*time.Time, error) {
	tagName := "release-" + version
	tagCommit, err := n.githubClient.GetTagCommit("nginx", "nginx", tagName)
	if err != nil {
		return nil, fmt.Errorf("could not get tags: %w", err)
	}

	return &tagCommit.Date, nil
}

func (n Nginx) getDependencySHA(dependencyURL, version string) (string, error) {
	nginxGPGKey, err := n.webClient.Get("http://nginx.org/keys/mdounin.key")
	if err != nil {
		return "", fmt.Errorf("could not get nginx GPG key: %w", err)
	}

	dependencySignature, err := n.webClient.Get(n.dependencySignatureURL(version))
	if err != nil {
		return "", fmt.Errorf("could not get dependency signature: %w", err)
	}

	dependencyOutputDir, err := ioutil.TempDir("", "nginx")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	dependencyOutputPath := filepath.Join(dependencyOutputDir, filepath.Base(dependencyURL))

	err = n.webClient.Download(dependencyURL, dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not download dependency: %w", err)
	}

	err = n.checksummer.VerifyASC(string(dependencySignature), dependencyOutputPath, string(nginxGPGKey))
	if err != nil {
		return "", fmt.Errorf("dependency signature verification failed: %w", err)
	}

	dependencySHA, err := n.checksummer.GetSHA256(dependencyOutputPath)
	if err != nil {
		return "", fmt.Errorf("could not get SHA256: %w", err)
	}

	return dependencySHA, nil
}

func (n Nginx) dependencySignatureURL(version string) string {
	return fmt.Sprintf("http://nginx.org/download/nginx-%s.tar.gz.asc", version)
}

func (n Nginx) dependencyURL(version string) string {
	return fmt.Sprintf("http://nginx.org/download/nginx-%s.tar.gz", version)
}
