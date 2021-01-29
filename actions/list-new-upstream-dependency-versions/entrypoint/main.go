package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	"os"
)

func main() {
	var (
		githubToken string
		name        string
	)

	flag.StringVar(&githubToken, "github-token", "", "Github access token")
	flag.StringVar(&name, "name", "", "Dependency name")
	flag.Parse()

	if name == "" {
		fmt.Println("`name` is required")
		os.Exit(1)
	}

	output, err := getNewVersions(githubToken, name)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}

	fmt.Println(output)
}

func getNewVersions(githubToken, name string) (string, error) {
	dep, err := dependency.NewDependencyFactory(githubToken).NewDependency(name)
	if err != nil {
		return "", fmt.Errorf("failed to create dependency: %w", err)
	}

	versions, err := dep.GetAllVersionRefs()
	if err != nil {
		return "", fmt.Errorf("failed to get versions: %w", err)
	}

	newVersionsJSON, err := json.Marshal(versions)
	if err != nil {
		return "", fmt.Errorf("failed to marshal new versions: %w", err)
	}

	return string(newVersionsJSON), nil
}
