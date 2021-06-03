package utils

import (
	"crypto/md5"
	"fmt"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	factory   dependency.DepFactory
	webClient PHPExtensionsWebClient
}

func NewPHPExtensionsUtils(factory dependency.DepFactory, webClient PHPExtensionsWebClient) PHPExtensionsUtils {
	return PHPExtensionsUtils{factory: factory, webClient: webClient}
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

	sort.Strings(versions)

	latestVersion := versions[len(versions)-1]

	depVersion, err := dep.GetDependencyVersion(latestVersion)
	if err != nil {
		return ExtensionVersion{}, fmt.Errorf("failed to get version '%s': %w", latestVersion, err)
	}

	err = p.webClient.DownloadExtensionsSource(depVersion.URI, depName)
	if err != nil {
		return ExtensionVersion{}, fmt.Errorf("failed to download extension %s source: %w", depName, err)
	}

	MD5, err := p.GetMD5(filepath.Join("/tmp", depName))
	if err != nil {
		return ExtensionVersion{}, err
	}

	return ExtensionVersion{
		Name:    depName,
		Version: depVersion.Version,
		MD5:     MD5,
	}, nil

	//output, err := json.Marshal(depVersion)
	//if err != nil {
	//	return "", fmt.Errorf("failed to marshal dependency version: %w", err)
	//}
	// create dep factory
	// get all the version refs (tested) and sort
	// call getdepversion (tested) with the latest version
	// download dep (tested) using the URL from depversion
	// calculate MD5 (tested)
	// return ExtensionVersion
}

func (p PHPExtensionsUtils) GenerateJSONPayload(folder string) (string, error) {

	//extensionVersion, err := GetLatestUpstreamVersion(dep,token)
	return "", nil
}

func (p PHPExtensionsUtils) GetMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "nil", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "nil", fmt.Errorf("failed to calculate MD5: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
