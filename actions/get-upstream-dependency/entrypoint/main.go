package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
)

func main() {
	var (
		githubToken string
		name        string
		version     string
	)

	flag.StringVar(&githubToken, "github-token", "", "Github access token")
	flag.StringVar(&name, "name", "", "Dependency name")
	flag.StringVar(&version, "version", "", "Dependency version")
	flag.Parse()

	if name == "" || version == "" {
		fmt.Println("`name` and `version` are required")
		os.Exit(1)
	}

	output, err := getDepVersion(githubToken, name, version)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}

	fmt.Println(output)
}

func getDepVersion(githubToken, name, version string) (string, error) {
	dep, err := dependency.NewDependencyFactory(githubToken).NewDependency(name)
	if err != nil {
		return "", fmt.Errorf("failed to create dependency: %w", err)
	}

	depVersion, err := dep.GetDependencyVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to get version '%s': %w", version, err)
	}

	output, err := json.Marshal(depVersion)
	if err != nil {
		return "", fmt.Errorf("failed to marshal dependency version: %w", err)
	}

	return string(output), nil
}
