package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/Masterminds/semver"
)

func main() {
	var (
		version string
	)

	flag.StringVar(&version, "version", "", "Dependency version")
	flag.Parse()

	if version == "" {
		fmt.Println("`version` is required")
		os.Exit(1)
	}

	// Strip the any non-digit prefix off the version and ensure that the new
	// version is semver-compatible
	reg, err := regexp.Compile(`([0-9]+.?)+`)
	if err != nil {
		fmt.Println("could not compile regexp for conversion")
		os.Exit(1)
	}

	semanticVersion, err := semver.NewVersion(reg.FindAllString(version, 1)[0])
	if err != nil {
		fmt.Println("could not convert to semantic version")
		os.Exit(1)
	}

	fmt.Print(semanticVersion.String())
}
