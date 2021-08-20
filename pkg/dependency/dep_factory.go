package dependency

import (
	"fmt"
	"time"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/internal"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/licenses"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency/purl"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Dependency
type Dependency interface {
	GetAllVersionRefs() ([]string, error)
	GetDependencyVersion(version string) (DepVersion, error)
	GetReleaseDate(version string) (*time.Time, error)
}

type DepVersion struct {
	Version         string     `json:"version"`
	URI             string     `json:"uri"`
	SHA256          string     `json:"sha256"`
	ReleaseDate     *time.Time `json:"release_date,omitempty"`
	DeprecationDate *time.Time `json:"deprecation_date,omitempty"`
	CPE             string     `json:"cpe"`
	PURL            string     `json:"purl"`
	Licenses        []string   `json:"licenses"`
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Checksummer
type Checksummer interface {
	VerifyASC(asc, path string, pgpKeys ...string) error
	VerifyMD5(path, md5 string) error
	VerifySHA1(path, sha string) error
	VerifySHA256(path, sha string) error
	VerifySHA512(path, sha string) error
	GetSHA256(path string) (string, error)
	SplitPGPKeys(block string) []string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . LicenseRetriever
type LicenseRetriever interface {
	LookupLicenses(dependencyName, sourceURL string) ([]string, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PURLGenerator
type PURLGenerator interface {
	Generate(dependencyName, dependencyVersion, sourceSHA, sourceURL string) string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . FileSystem
type FileSystem interface {
	WriteFile(filename, contents string) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GithubClient
type GithubClient interface {
	GetReleaseTags(org, repo string) ([]internal.GithubRelease, error)
	GetTags(org, repo string) ([]string, error)
	GetReleaseAsset(org, repo, version, filename string) ([]byte, error)
	DownloadReleaseAsset(org, repo, version, filename, outputPath string) (url string, err error)
	DownloadSourceTarball(org, repo, version, outputPath string) (url string, err error)
	GetTagCommit(org, repo, version string) (internal.GithubTagCommit, error)
	GetReleaseDate(org, repo, tag string) (*time.Time, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . WebClient
type WebClient interface {
	Download(url, outputPath string, options ...internal.RequestOption) error
	Get(url string, options ...internal.RequestOption) ([]byte, error)
}

type DepFactory struct {
	checksummer      Checksummer
	fileSystem       FileSystem
	githubClient     GithubClient
	webClient        WebClient
	licenseRetriever LicenseRetriever
	purlGenerator    PURLGenerator
}

func NewCustomDependencyFactory(checksum Checksummer, fileSystem FileSystem, githubClient GithubClient, webClient WebClient, licenseRetriever LicenseRetriever, purlGenerator PURLGenerator) DepFactory {
	return DepFactory{
		checksummer:      checksum,
		fileSystem:       fileSystem,
		githubClient:     githubClient,
		webClient:        webClient,
		licenseRetriever: licenseRetriever,
		purlGenerator:    purlGenerator,
	}
}

func NewDependencyFactory(accessToken string) DepFactory {
	checksummer := internal.NewChecksummer()
	fileSystem := internal.NewFileSystem()
	webClient := internal.NewWebClient()
	githubClient := internal.NewGithubClient(webClient, accessToken)
	licenseRetriever := licenses.NewLicenseRetriever()
	purlGenerator := purl.NewPURLGenerator()

	return DepFactory{
		checksummer:      checksummer,
		fileSystem:       fileSystem,
		githubClient:     githubClient,
		webClient:        webClient,
		licenseRetriever: licenseRetriever,
		purlGenerator:    purlGenerator,
	}
}

func (d DepFactory) SupportsDependency(name string) bool {
	_, err := d.NewDependency(name)

	if err != nil {
		return false
	}

	return true
}

func (d DepFactory) NewDependency(name string) (Dependency, error) {
	switch name {
	case "apc", "apcu":
		return Pecl{
			productName:      name,
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "bundler":
		return Bundler{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "composer":
		return Composer{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			githubClient:     d.githubClient,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "curl":
		return Curl{
			checksummer:      d.checksummer,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "dotnet-aspnetcore":
		return DotnetASPNETCore{
			checksummer:      d.checksummer,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "dotnet-runtime":
		return DotnetRuntime{
			checksummer:      d.checksummer,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "dotnet-sdk":
		return DotnetSDK{
			checksummer:      d.checksummer,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "go":
		return Go{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "httpd":
		return Httpd{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "icu":
		return ICU{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			githubClient:     d.githubClient,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "nginx":
		return Nginx{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			githubClient:     d.githubClient,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "node":
		return Node{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "php":
		return Php{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "pip", "pipenv":
		return PyPi{
			productName:      name,
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "python":
		return Python{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "ruby":
		return Ruby{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "rust":
		return Rust{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			githubClient:     d.githubClient,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "tini":
		return Tini{
			checksummer:      d.checksummer,
			githubClient:     d.githubClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	case "yarn":
		return Yarn{
			checksummer:      d.checksummer,
			fileSystem:       d.fileSystem,
			githubClient:     d.githubClient,
			webClient:        d.webClient,
			licenseRetriever: d.licenseRetriever,
			purlGenerator:    d.purlGenerator,
		}, nil
	default:
		return nil, fmt.Errorf("dependency type '%s' is not supported", name)
	}
}
