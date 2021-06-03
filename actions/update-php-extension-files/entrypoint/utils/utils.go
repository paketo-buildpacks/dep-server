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

func (p PHPExtensionsUtils) getUpdatedExtensions(file PHPExtMetadataFile) ([]PHPExtension, err) {
	for _, dep

	nativeModules := 
	extensions := 
}

func (p PHPExtensionsUtils) GenerateJSONPayload(folder string) (string, error) {

	ymlFiles := getPHPExtensionsYMLFiles(folder)

	for _, ymlFile := range ymlFiles {
		updatedExtensions := p.getUpdatedExtensions(ymlFile.Nati)
	}
	// Walk folder to find php extension ymls
	// Parse each yml
	// For each extension and native module, get the latest upstream version
	// compare the latest upstream to the current extension version

	// If latest is greater, add it as an ExtensionVersion to a struct in the
	// following format:

	// type Payload struct {
	//   Data map[string][]PHPExtMetadataFile
	//  }

	// Example:
	// p.Data["php-8-yml"] = append(p.Data["php-8-yml"], newExtVersion)

	// Marshal into JSON, return JSON as string or empty string and an error

	//extensionVersion, err := GetLatestUpstreamVersion(dep,token)

	return "", nil
}

func (p PHPExtensionsUtils) getPHPExtensionsYMLFiles(folder string) ([]PHPExtMetadataFile, error) {

	var (
		YMLFiles []PHPExtMetadataFile
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
			YMLFiles = append(YMLFiles, YMLFile)
		}

		return nil
	})

	if err != nil {
		return []PHPExtMetadataFile{}, err
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
