package main

import (
	"os"

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

func ParseYML(YMLFile string) (PHPExtMetadataFile, error) {

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

func GetLatestUpstreamVersion(depName string) (ExtensionVersion, error) {

	// create dep factory
	// get all the version refs (tested) and sort
	// call getdepversion (tested) with the latest version
	// download dep (tested) using the URL from depversion
	// calculate MD5 (tested)
	// return ExtensionVersion

	return ExtensionVersion{}, nil
}
