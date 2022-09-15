package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

type Options struct {
	BuildpackDir string
	Version      string
	ArtifactPath string
}

func main() {
	var options Options
	flag.StringVar(&options.BuildpackDir, "buildpack-dir", "", "Absolute path to buildpack directory")
	flag.StringVar(&options.Version, "version", "1.2.3", "OPTIONAL, version to compile and/or test (if applicable)")
	flag.StringVar(&options.ArtifactPath, "artifact-path", "", "OPTIONAL, absolute path to a local artifact to run `make test` against (if applicable).\n If not provided and compilation runs, compiled tarball will be used for testing.")
	flag.Parse()

	if options.BuildpackDir == "" {
		flag.Usage()
		log.Fatal(fmt.Errorf("missing required flag buildpack-dir"))
	}

	// check for the existence of a dependency directory
	if _, err := os.Stat(filepath.Join(options.BuildpackDir, "dependency")); os.IsNotExist(err) {
		printError(fmt.Sprintf("Error: No `dependency` directory found in %s. Did you pass in an absolute path?", options.BuildpackDir))
		return
	}

	printSuccessful("✅ `dependency` directory found\n")

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		printError("Error: could not make temporary directory for testing")
	}

	// SKIP testing if the user did not pass in an artifact/version to test, and compilation did not produce an artifact

	retrieveErr := retrieve(options.BuildpackDir, tmpDir)
	artifact, compileErr := compile(options.BuildpackDir, tmpDir, options.Version)

	if options.ArtifactPath != "" {
		fmt.Printf("Artifact path set by user, %s will be used for testing", options.ArtifactPath)
		artifact = options.ArtifactPath
	}

	testErr := test(options.BuildpackDir, options.Version, artifact)

	fmt.Print("\nCleaning up files\n\n")
	cleanup(tmpDir)

	if retrieveErr != nil && compileErr != nil && testErr != nil {
		printError("❌ Validation failed")
	} else {
		printSuccessful("All checks passed")
	}
}

func compile(buildpackDir, tmpDir, version string) (string, error) {
	compilationDir, err := filepath.Abs(filepath.Join(buildpackDir, "dependency", "actions", "compile"))
	if err != nil {
		printError(err.Error())
		return "", err
	}

	if _, err := os.Stat(compilationDir); os.IsNotExist(err) {
		fmt.Printf("\n2/3: Skipping `compile` checks. No `compile` action found at %s.\n\n", compilationDir)
		return "", nil
	}

	fmt.Print("\n2/3: Validating Compilation action\n")

	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Unable to create docker client")
		printError(err.Error())
		return "", err
	}

	// get a target
	dockerfiles, err := filepath.Glob(filepath.Join(compilationDir, "*Dockerfile"))
	if err != nil || len(dockerfiles) < 1 {
		printError(fmt.Sprintf("Error: could not find a Dockerfile in the compilation action as required: %v", err))
		return "", err
	}

	// just test against the first Dockerfile found
	dockerfile := filepath.Base(dockerfiles[0])
	target := strings.Split(dockerfile, ".")[0]
	fmt.Println(target)

	fmt.Printf("Building %s compilation image...\n", target)
	dockerBuildContext, err := archive.TarWithOptions(compilationDir, &archive.TarOptions{})
	if err != nil {
		printError(err.Error())
		return "", err
	}
	defer dockerBuildContext.Close()

	resp, err := cli.ImageBuild(context.Background(), dockerBuildContext, types.ImageBuildOptions{
		Dockerfile:     dockerfile,
		SuppressOutput: false,
		NoCache:        true,
		Tags:           []string{"compilation"},
		Remove:         true,
	})
	if err != nil {
		printError(err.Error())
		return "", err
	}
	defer resp.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	err = jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)
	if err != nil {
		printError(err.Error())
		return "", err
	}

	fmt.Println("\nCreating Container")
	resp2, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: "compilation",
		Cmd:   []string{"--version", version, "--outputDir", "/output", "--target", target},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: tmpDir,
				Target: "/output",
			},
		},
	},
		nil, nil, "")
	if err != nil {
		printError(err.Error())
		return "", err
	}

	fmt.Println("Running compilation")
	if err := cli.ContainerStart(context.Background(), resp2.ID, types.ContainerStartOptions{}); err != nil {
		printError(err.Error())
		return "", err
	}

	statusCh, errCh := cli.ContainerWait(
		context.Background(),
		resp2.ID,
		container.WaitConditionNotRunning,
	)
	select {
	case err := <-errCh:
		if err != nil {
			printError(err.Error())
			return "", err
		}
	case status := <-statusCh:
		log.Printf("Compilation container exited with status code: %#v\n\n", status.StatusCode)
	}

	artifact, err := filepath.Glob(tmpDir + "/" + "*" + ".tgz")
	if err != nil || len(artifact) <= 0 {
		printError(fmt.Sprintf("Error: no artifact with name *-%s-%s.tgz found at %s", version, target, tmpDir))
		return "", err
	}

	file := artifact[0]
	printSuccessful(fmt.Sprintf("✅ compilation has successfully created %s", file))

	shaFile, err := filepath.Glob(tmpDir + "/" + "*" + ".tgz.sha256")
	if err != nil || len(shaFile) <= 0 {
		printError(fmt.Sprintf("Error: no artifact with name *-%s-%s.tgz found at %s", version, target, tmpDir))
		return "", err
	}

	for _, file := range shaFile {
		printSuccessful(fmt.Sprintf("✅ compilation has successfully created %s", file))
	}

	containerCleanup(cli, resp2.ID)
	return file, nil
}

func retrieve(buildpackDir, tmpDir string) error {
	args := []string{"retrieve", fmt.Sprintf("buildpackTomlPath=%s/buildpack.toml", buildpackDir), fmt.Sprintf("output=%s/metadata.json", tmpDir)}
	fmt.Println("1/3: Validating Retrieval")
	fmt.Printf("Running `make %s`\n", strings.Join(args, " "))

	e := exec.Command("make", args...)

	// Run from the dependency directory
	var err error
	e.Dir, err = filepath.Abs(filepath.Join(buildpackDir, "dependency"))
	if err != nil {
		printError("Error: could not get the absolute path to dependency directory")
		return err
	}

	var out bytes.Buffer
	e.Stdout = &out
	e.Stderr = &out
	err = e.Run()
	if err != nil {
		if strings.Contains(out.String(), "No rule to make target `retrieve'.") {
			printError(fmt.Sprintf("No Makefile containing target `retrieve`: %v\n", err))
			return err
		}
		if strings.Contains(out.String(), "[retrieve] Error") {
			fmt.Println(out)
			printError(fmt.Sprintf("`retrieve` did not handle expected inputs correctly: %v\n", err))
			return err
		}
	}

	// validate metadata.json is created
	if _, err = os.Stat(filepath.Join(tmpDir, "metadata.json")); os.IsNotExist(err) {
		printError(fmt.Sprintf("Error: metadata.json file not created at %s/metadata.json as expected.", tmpDir))
		return err
	}

	printSuccessful("✅ retrieve has successfully created a metadata.json")

	// if there is a compilation action, the SHA256 and URI fields in the metadata should be empty.
	includeSHAandURI := false

	compilationDir := filepath.Join(buildpackDir, "dependency", "actions", "compile")
	if _, err := os.Stat(compilationDir); os.IsNotExist(err) {
		includeSHAandURI = true
	}

	validate := validateMetadata(filepath.Join(tmpDir, "metadata.json"), includeSHAandURI)
	return validate
}

func validateMetadata(metadataFile string, includeSHAandURI bool) error {
	type configStruct struct {
		CPE             *string        `json:"cpe,omitempty"`
		PURL            *string        `json:"purl,omitempty"`
		DeprecationDate *time.Time     `json:"deprecation_date,omitempty"`
		ID              *string        `json:"id,omitempty"`
		Licenses        *[]interface{} `json:"licenses,omitempty"`
		Name            *string        `json:"name,omitempty"`
		SHA256          *string        `json:"sha256,omitempty"`
		Source          *string        `json:"source,omitempty"`
		SourceSHA256    *string        `json:"source_sha256,omitempty"`
		Stacks          *[]string      `json:"stacks,omitempty"`
		URI             *string        `json:"uri,omitempty"`
		Version         *string        `json:"version,omitempty"`
		Target          *string        `json:"target,omitempty"`
	}

	file, err := os.Open(metadataFile)
	if err != nil {
		printError(fmt.Sprintf("Error: failed to open metadata file: %v", err))
		return err
	}

	var config []configStruct
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println(" ⚠️  Warning: No metadata generated to decode. You may want to remove a version from the buildpack.toml and re-run validation so metadata can be checked.")
		printError(fmt.Sprintf("Error: failed to decode metadata file: %v", err))
		return err
	}

	if len(config) > 0 {
		if config[0].CPE == nil {
			fmt.Println(" ⚠️  Warning: metadata missing CPE field")
		}
		if config[0].PURL == nil {
			fmt.Println(" ⚠️  Warning: metadata missing PURL field")
		}
		if config[0].DeprecationDate == nil {
			fmt.Println(" ⚠️  Warning: metadata missing DeprecationDate field")
		}
		if config[0].ID == nil {
			fmt.Println(" ⚠️  Warning: metadata missing ID field")
		}
		if config[0].Licenses == nil {
			fmt.Println(" ⚠️  Warning: metadata missing Licenses field")
		}
		if config[0].Name == nil {
			fmt.Println(" ⚠️  Warning: metadata missing Name field")
		}
		if config[0].Source == nil {
			fmt.Println(" ⚠️  Warning: metadata missing Source field")
		}
		if config[0].SourceSHA256 == nil {
			fmt.Println(" ⚠️  Warning: metadata missing SourceSHA256 field")
		}
		if config[0].Stacks == nil {
			fmt.Println(" ⚠️  Warning: metadata missing Stacks field")
		}
		if config[0].Target == nil {
			fmt.Println(" ⚠️  Warning: metadata missing Target field")
		}
		if config[0].Version == nil {
			fmt.Println(" ⚠️  Warning: metadata missing Version field")
		}
		if config[0].URI == nil && includeSHAandURI {
			fmt.Println(" ⚠️  Warning: metadata missing URI field")
		}
		if config[0].SHA256 == nil && includeSHAandURI {
			fmt.Println(" ⚠️  Warning: metadata missing SHA256 field")
		}
	} else {
		fmt.Println(" ⚠️  Warning: generated metadata is blank. You may want to remove a version from the buildpack.toml and re-run validation so metadata can be checked.")
	}

	if includeSHAandURI == false {
		if len(config) > 0 {
			for _, entry := range config {
				if entry.SHA256 != nil || entry.URI != nil {
					printError("Error: there is compilation code available, but the SHA256 and/or URI metadata field is populated. If the dependency is to be compiled, these fields should both be blank.")
					return err
				}
			}
		}
	}
	printSuccessful("✅ overall output metadata structure is correct\n")
	return nil
}

func printSuccessful(content string) {
	colorGreen := "\033[32m"
	colorReset := "\033[0m"
	fmt.Println(string(colorGreen), content, string(colorReset))
}

func printError(content string) {
	colorRed := "\033[31m"
	colorReset := "\033[0m"
	fmt.Println(string(colorRed), content, string(colorReset))
}

func cleanup(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Warning: could not remove %s\n", path)
	}
}

func containerCleanup(cli *client.Client, containerID string) {
	fmt.Print("Cleaning up containers\n\n")
	var err error

	if err := cli.ContainerStop(context.Background(), containerID, nil); err != nil {
		fmt.Printf("⚠️  Warning: Unable to stop container %s: %s\n", containerID, err)
	}

	if err = cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		fmt.Printf("⚠️  Warning: Unable to remove container %s: %s\n", containerID, err)
	}

	_, err = cli.ImageRemove(context.Background(), "compilation", types.ImageRemoveOptions{PruneChildren: true})
	if err != nil {
		fmt.Printf("⚠️  Warning: Unable to remove image `compilation`: %s\n", err)
	}
}

func test(buildpackDir, version, artifactPath string) error {
	testDir := filepath.Join(buildpackDir, "dependency", "test")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		fmt.Printf("\n3/3: Skipping `test` checks. No `test` directory found in %s.\n", testDir)
		return nil
	}

	if version == "" || artifactPath == "" {
		fmt.Println("\n3/3: Skipping `test` checks. No `--version` and `--artifact-path` passed in and/or compilation didn't produce an artifact.")
		fmt.Println("       User should manually check that `make test` takes in the correct inputs, or pass `--version` and `--artifact-path` to the validator")
		return nil
	}

	args := []string{"test", fmt.Sprintf("version=%s", version), fmt.Sprintf("tarballPath=%s", artifactPath)}
	fmt.Printf("\n3/3: Validating `make %s`\n", strings.Join(args, " "))

	e := exec.Command("make", args...)
	e.Dir = filepath.Join(buildpackDir, "dependency")

	var out bytes.Buffer
	e.Stdout = &out
	e.Stderr = &out
	err := e.Run()
	if err != nil {
		if strings.Contains(out.String(), "No rule to make target `test'.") {
			printError(fmt.Sprintf("No Makefile containing target `test`: %v\n", err))
		}
		fmt.Println(out.String())
		printError(fmt.Sprintf("Error: `make test` failed: %v\n", err))
		return err
	}

	fmt.Println(out.String())
	printSuccessful("✅ `test` ran successfully\n")
	return nil
}
