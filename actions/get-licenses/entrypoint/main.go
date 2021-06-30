package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/paketo-buildpacks/packit/vacation"
)

type Options struct {
	DependencyName string
	URL            string
}

func main() {
	var options Options

	flag.StringVar(&options.DependencyName, "dependency-name", "", "Dependency name")
	flag.StringVar(&options.URL, "url", "", "Dependency source URL")
	flag.Parse()

	requiredFlags := map[string]string{
		"--dependency-name": options.DependencyName,
		"--url":             options.URL,
	}

	for name, value := range requiredFlags {
		if value == "" {
			fail(fmt.Errorf("missing required flag %s", name))
		}
	}

	// CAAPM dependency does not have an auto-retrievable license
	// Exit the function and leave the license blank
	if options.DependencyName == "CAAPM" || options.DependencyName == "composer" {
		fmt.Printf("Skipping license retrieval for %s\n", options.DependencyName)
		fmt.Println("License is not automatically retrievable and may need to be looked up manually")
		fmt.Printf("::set-output name=licenses::%v\n", []string{})

		os.Exit(0)
	}

	fmt.Printf("Getting the dependency artifact from %s\n", options.URL)
	url := options.URL
	resp, err := http.Get(url)
	if err != nil {
		fail(fmt.Errorf("failed to query url: %w", err))
	}
	if resp.StatusCode != http.StatusOK {
		fail(fmt.Errorf("failed to query url %s with: status code %d", url, resp.StatusCode))
	}

	fmt.Println("Decompressing the dependency artifact")
	tempDir, err := os.MkdirTemp("", "destination")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(tempDir)

	switch options.DependencyName {
	case "bundler":
		err := bundlerDecompress(resp.Body, tempDir)
		if err != nil {
			fail(err)
		}
	case "dotnet-runtime", "dotnet-aspnetcore", "dotnet-sdk":
		err := defaultDecompress(resp.Body, tempDir, 0)
		if err != nil {
			fail(err)
		}
	default:
		err := defaultDecompress(resp.Body, tempDir, 1)
		if err != nil {
			fail(err)
		}
	}

	fmt.Println("Scanning artifact for license file")
	filer, err := filer.FromDirectory(tempDir)
	if err != nil {
		fail(fmt.Errorf("failed to setup a licensedb filer: %w", err))
	}

	licenses, err := licensedb.Detect(filer)
	if err != nil {
		fail(fmt.Errorf("failed to detect licenses: %w", err))
	}

	// Only return the license IDs, in alphabetical order
	var licenseIDs []string
	for key := range licenses {
		licenseIDs = append(licenseIDs, key)
	}
	sort.Strings(licenseIDs)

	fmt.Printf("::set-output name=licenses::%v\n", licenseIDs)
	fmt.Println("Licenses found!")
}

func defaultDecompress(artifact io.Reader, destination string, stripComponents int) error {
	archive := vacation.NewArchive(artifact)

	err := archive.StripComponents(stripComponents).Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress source file: %w", err)
	}

	return nil
}

// The bundler dependency comes as a .gem file (tar.gz mime type) with a
// data.tar.gz file inside that contains the license.
func bundlerDecompress(artifact io.Reader, destination string) error {
	archive := vacation.NewArchive(artifact)
	err := archive.Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress source file: %w", err)
	}

	innerArtifact, _ := os.Open(filepath.Join(destination, "data.tar.gz"))
	innerArchive := vacation.NewArchive(innerArtifact)
	err = innerArchive.Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress inner source file: %w", err)
	}

	return nil
}

func fail(err error) {
	fmt.Printf("Error: %s", err)
	os.Exit(1)
}
