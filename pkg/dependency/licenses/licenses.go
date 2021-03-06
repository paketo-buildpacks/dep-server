package licenses

import (
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

type LicenseRetriever struct{}

func NewLicenseRetriever() LicenseRetriever {
	return LicenseRetriever{}
}

func (LicenseRetriever) LookupLicenses(dependencyName, sourceURL string) ([]string, error) {
	// composer dependency does not have an auto-retrievable license
	// Exit the function and leave the license blank
	if dependencyName == "composer" {
		// skipping license retrieval
		// license is not automatically retrievable and may need to be looked up manually
		return []string{}, nil
	}

	// getting the dependency artifact from sourceURL
	url := sourceURL
	resp, err := http.Get(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to query url: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("failed to query url %s with: status code %d", url, resp.StatusCode)
	}

	// decompressing the dependency artifact
	tempDir, err := os.MkdirTemp("", "destination")
	if err != nil {
		return []string{}, err
	}
	defer os.RemoveAll(tempDir)

	switch dependencyName {
	case "bundler":
		err := bundlerDecompress(resp.Body, tempDir)
		if err != nil {
			return []string{}, err
		}
	case "dotnet-runtime", "dotnet-aspnetcore", "dotnet-sdk":
		err := defaultDecompress(resp.Body, tempDir, 0)
		if err != nil {
			return []string{}, err
		}
	default:
		err := defaultDecompress(resp.Body, tempDir, 1)
		if err != nil {
			return []string{}, err
		}
	}

	// scanning artifact for license file
	filer, err := filer.FromDirectory(tempDir)
	if err != nil {
		return []string{}, fmt.Errorf("failed to setup a licensedb filer: %w", err)
	}

	licenses, err := licensedb.Detect(filer)
	// if no licenses are found, just return an empty slice.
	if err != nil {
		if err.Error() != "no license file was found" {
			return []string{}, fmt.Errorf("failed to detect licenses: %w", err)
		}
		return []string{}, nil
	}

	// Only return the license IDs, in alphabetical order
	var licenseIDs []string
	for key := range licenses {
		licenseIDs = append(licenseIDs, key)
	}
	sort.Strings(licenseIDs)

	return licenseIDs, nil
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
