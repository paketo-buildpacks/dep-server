package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"gopkg.in/yaml.v2"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PHPExtUtils
type PHPExtUtils interface {
	GetPHPExtensionsYMLFiles(folder string) (map[string]PHPExtMetadataFile, error)
	ParseYML(YMLFile string) (PHPExtMetadataFile, error)
	GetLatestUpstreamVersion(depName string) (ExtensionVersion, error)
	GetUpdatedMetadataFile(file PHPExtMetadataFile) (PHPExtMetadataFile, error)
	GetUpdatedExtensions(currentExtensions []PHPExtension) ([]PHPExtension, error)
}

type PHPExtension struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	MD5     string `yaml:"md5"`
	Klass   string `yaml:"klass"`
}

type PHPExtMetadataFile struct {
	NativeModules []PHPExtension `yaml:"native_modules"`
	Extensions    []PHPExtension `yaml:"extensions"`
}

type ExtensionVersion struct {
	Name    string
	Version string
	MD5     string
}

type PHPExtensionsUtils struct {
	factory     dependency.DepFactory
	webClient   PHPExtensionsWebClient
	checkSummer ChecksummerPHP
}

type JSONPayload struct {
	Data map[string]PHPExtMetadataFile `json:"data"`
}

func NewPHPExtensionsUtils(factory dependency.DepFactory, webClient PHPExtensionsWebClient, checkSummer ChecksummerPHP) PHPExtensionsUtils {
	return PHPExtensionsUtils{
		factory:     factory,
		webClient:   webClient,
		checkSummer: checkSummer,
	}
}

func (p PHPExtensionsUtils) ParseYML(YMLFile string) (PHPExtMetadataFile, error) {

	file, err := os.Open(YMLFile)
	if err != nil && !os.IsNotExist(err) {
		return PHPExtMetadataFile{}, err
	}

	defer file.Close()

	var phpMetadataFile PHPExtMetadataFile

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&phpMetadataFile)
		if err != nil {
			return PHPExtMetadataFile{}, err
		}
	}

	return phpMetadataFile, nil
}

func (p PHPExtensionsUtils) GetLatestUpstreamVersion(depName string) (ExtensionVersion, error) {

	dep, err := p.factory.NewDependency(depName)
	if err != nil {
		return ExtensionVersion{}, fmt.Errorf("failed to create dependency: %w", err)
	}

	versions, err := dep.GetAllVersionRefs()
	if err != nil {
		return ExtensionVersion{}, fmt.Errorf("failed to get versions: %w", err)
	}

	semverVersions := make([]*semver.Version, len(versions))
	for i, r := range versions {
		semverVersion, err := semver.NewVersion(r)
		if err != nil {
			return ExtensionVersion{}, fmt.Errorf("error parsing version: %s", err)
		}
		semverVersions[i] = semverVersion
	}

	sort.Sort(semver.Collection(semverVersions))

	latestVersion := semverVersions[len(semverVersions)-1].String()

	depVersion, err := dep.GetDependencyVersion(latestVersion)
	if err != nil {
		return ExtensionVersion{}, fmt.Errorf("failed to get version '%s': %w", latestVersion, err)
	}

	phpExtensionSource := filepath.Join("/tmp", filepath.Base(depVersion.URI))

	err = p.webClient.DownloadExtensionsSource(depVersion.URI, phpExtensionSource)
	if err != nil {
		return ExtensionVersion{}, fmt.Errorf("failed to download extension %s source: %w", depName, err)
	}

	MD5, err := p.GetMD5(phpExtensionSource)
	if err != nil {
		return ExtensionVersion{}, err
	}

	return ExtensionVersion{
		Name:    depName,
		Version: depVersion.Version,
		MD5:     MD5,
	}, nil
}

func (p PHPExtensionsUtils) GetUpdatedMetadataFile(file PHPExtMetadataFile) (PHPExtMetadataFile, error) {

	nativeModulesUpdated, err := p.GetUpdatedExtensions(file.NativeModules)
	if err != nil {
		return PHPExtMetadataFile{}, err
	}

	extensionsUpdated, err := p.GetUpdatedExtensions(file.Extensions)
	if err != nil {
		return PHPExtMetadataFile{}, err
	}

	return PHPExtMetadataFile{
		NativeModules: nativeModulesUpdated,
		Extensions:    extensionsUpdated,
	}, nil
}

func (p PHPExtensionsUtils) GetUpdatedExtensions(currentExtensions []PHPExtension) ([]PHPExtension, error) {

	var extensionsUpdated []PHPExtension

	for _, nativeModule := range currentExtensions {

		latestUpstreamVersion, err := p.GetLatestUpstreamVersion(nativeModule.Name)
		if err != nil {
			return []PHPExtension{}, err
		}

		actualSemverVersion, err := semver.NewVersion(nativeModule.Version)
		latestSemverVersion, err := semver.NewVersion(latestUpstreamVersion.Version)

		if actualSemverVersion.LessThan(latestSemverVersion) {
			extensionsUpdated = append(extensionsUpdated, PHPExtension{
				Name:    latestUpstreamVersion.Name,
				Version: latestUpstreamVersion.Version,
				MD5:     latestUpstreamVersion.MD5,
				Klass:   nativeModule.Klass,
			})
		}
	}

	return extensionsUpdated, nil
}

func (p PHPExtensionsUtils) GetPHPExtensionsYMLFiles(folder string) (map[string]PHPExtMetadataFile, error) {

	var (
		YMLFiles map[string]PHPExtMetadataFile
	)

	err := filepath.WalkDir(folder, func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), "php") {
			YMLFile, err := p.ParseYML(path)
			if err != nil {

				return err
			}
			YMLFiles[info.Name()] = YMLFile
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return YMLFiles, nil
}

func (p PHPExtensionsUtils) GetMD5(path string) (string, error) {
	md5, err := p.checkSummer.GetMD5(path)
	if err != nil {
		return "", err
	}
	defer os.Remove(path)

	return md5, nil
}
