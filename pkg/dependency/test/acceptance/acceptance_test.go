package acceptance_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency"
	derrors "github.com/paketo-buildpacks/dep-server/pkg/dependency/errors"
)

const githubAccessTokenEnvVar = "GITHUB_ACCESS_TOKEN"
const versionsToTest = 10

func TestAcceptance(t *testing.T) {
	if _, ok := os.LookupEnv(githubAccessTokenEnvVar); !ok {
		t.Fatalf("Must set %s", githubAccessTokenEnvVar)
	}

	spec.Run(t, "Acceptance", testAcceptance, spec.Report(report.Terminal{}))
}

func testAcceptance(t *testing.T, when spec.G, it spec.S) {
	var (
		assert       = assert.New(t)
		require      = require.New(t)
		dependencies = []string{
			"apcu",
			"bundler",
			"CAAPM",
			"composer",
			"curl",
			"dotnet-aspnetcore",
			"dotnet-runtime",
			"dotnet-sdk",
			"go",
			"httpd",
			"icu",
			"nginx",
			"node",
			"php",
			"pip",
			"pipenv",
			"python",
			"ruby",
			"rust",
			"tini",
			"yarn",
		}
	)

	when("listing and getting versions", func() {
		for _, depName := range dependencies {
			func(depName string) {
				it(fmt.Sprintf("lists all versions and fetches the oldest and newest %d versions of %s", versionsToTest, depName), func() {
					dep, err := dependency.NewDependencyFactory(os.Getenv(githubAccessTokenEnvVar)).NewDependency(depName)
					require.NoError(err)

					versions, err := dep.GetAllVersionRefs()
					require.NoError(err, "error listing versions of %s", depName)
					assert.NotEmpty(versions)

					var depVersions []dependency.DepVersion
					for _, i := range versionIndicesToGet(versions) {
						version := versions[i]

						depVersion, err := dep.GetDependencyVersion(version)
						if err != nil {
							var noSourceCodeError derrors.NoSourceCodeError
							if errors.As(err, &noSourceCodeError) {
								continue
							}
							require.NoError(err, "error getting %s %s", depName, version)
						}

						depVersions = append(depVersions, depVersion)

						assert.Equal(version, depVersion.Version)
						assert.Len(depVersion.SHA256, 64, "SHA256 did not have 64 characters for %s %s", depName, version)
						assert.NotEmpty(depVersion.URI)

						parsedVersion, err := semver.NewVersion(strings.TrimPrefix(version, "go"))
						if err == nil {
							assert.Empty(parsedVersion.Prerelease())
						} else {
							fmt.Printf("WARNING: could not parse version '%s' of %s as semver, skipping final version check\n", version, depName)
						}

						if depName != "CAAPM" {
							assert.False(depVersion.ReleaseDate.IsZero())
						}
					}

					if depName != "CAAPM" {
						if depName == "bundler" || depName == "composer" || depName == "nginx" {
							for i := 0; i < len(depVersions)-1; i++ {
								d := 24 * time.Hour
								secondDay := depVersions[i+1].ReleaseDate.Truncate(d)
								assert.True(depVersions[i].ReleaseDate.Truncate(d).After(secondDay) || depVersions[i].ReleaseDate.Truncate(d).Equal(secondDay))
							}
						} else {
							for i := 0; i < len(depVersions)-1; i++ {
								assert.True(depVersions[i].ReleaseDate.After(*depVersions[i+1].ReleaseDate) || depVersions[i].ReleaseDate.Equal(*depVersions[i+1].ReleaseDate), fmt.Sprintf("failed with %s and %s", depVersions[i], depVersions[i+1]))
							}
						}
					}
				})
			}(depName)
		}
	}, spec.Parallel())
}

func versionIndicesToGet(versions []string) []int {
	newestVersionsStartIndex := 0
	newestVersionsEndIndex := versionsToTest - 1
	oldestVersionsStartIndex := len(versions) - versionsToTest
	oldestVersionsEndIndex := len(versions) - 1

	if newestVersionsEndIndex >= len(versions)-1 {
		newestVersionsEndIndex = len(versions) - 1
	}

	if oldestVersionsStartIndex <= newestVersionsEndIndex {
		oldestVersionsStartIndex = newestVersionsEndIndex + 1
	}

	var indices []int
	for i := newestVersionsStartIndex; i <= newestVersionsEndIndex; i++ {
		indices = append(indices, i)
	}
	for i := oldestVersionsStartIndex; i <= oldestVersionsEndIndex; i++ {
		indices = append(indices, i)
	}

	return indices
}
