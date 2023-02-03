package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

type Stack struct {
	ID                string   `json:"id"`
	VersionConstraint string   `json:"version-constraint,omitempty"`
	Mixins            []string `json:"mixins,omitempty"`
}

func main() {
	var config struct {
		Version    string
		StacksJSON string
	}

	flag.StringVar(&config.StacksJSON,
		"stacks",
		"",
		"JSON array of stack IDs with version constraints")
	flag.StringVar(&config.Version,
		"version",
		"",
		"Version of dependency")

	flag.Parse()

	if config.StacksJSON == "" {
		config.StacksJSON = `[]`
	}

	var stacks []Stack
	err := json.Unmarshal([]byte(config.StacksJSON), &stacks)
	if err != nil {
		log.Fatal(err)
	}

	var result []Stack
	for _, stack := range stacks {
		if stack.VersionConstraint == "" {
			result = append(result, Stack{ID: stack.ID, Mixins: stack.Mixins})
			continue
		}
		c, err := semver.NewConstraint(stack.VersionConstraint)
		if err != nil {
			log.Fatal(err)
		}
		v, _ := semver.NewVersion(config.Version)
		if err != nil {
			log.Fatal(err)
		}
		compatible, msgs := c.Validate(v)
		if compatible {
			result = append(result, Stack{ID: stack.ID, Mixins: stack.Mixins})
			continue
		}
		for _, m := range msgs {
			log.Println(m)
		}
	}
	output, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	if len(result) == 0 {
		log.Fatal("No stacks compatible with this version")
	}

	log.Println("Output: ", string(output))
	actions.Outputs{
		"compatible-stacks": string(output),
	}.Write()
}
