package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/paketo-buildpacks/dep-server/pkg/dependency/licenses"
)

type DepMetadata struct {
	Version         string `json:"version"`
	URI             string `json:"uri"`
	SHA256          string `json:"sha256"`
	Source          string `json:"source"`
	SourceSHA256    string `json:"source_sha256"`
	DeprecationDate string `json:"deprecation_date"`
	CPE             string `json:"cpe,omitempty"`
	Licenses        string `json:"licenses"`
}

func main() {
	// Takes in the name of 1 dep => dispatches to the test-upload workflow with all metadata
	dependencyName := "tini"

	// reach out to the api.deps..../<dep-name> get all the metadata for all versions
	resp, err := http.Get(fmt.Sprintf("https://api.deps.paketo.io/v1/dependency?name=%s", dependencyName))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// translate JSON
	var deps []DepMetadata
	err = json.NewDecoder(resp.Body).Decode(&deps)
	if err != nil {
		log.Fatal(err)
	}

	// for each dep version ... get all metadata except licenses
	licenseRetriever := licenses.NewLicenseRetriever()
	for _, dep := range deps {
		// pass the dep name and source URL and whatever else to pkg/dependency/licenses to get licenses
		licenses, err := licenseRetriever.LookupLicenses(dependencyName, dep.Source)
		if err != nil {
			log.Fatal(err)
		}

		licensePayload, err := json.Marshal(licenses)
		if err != nil {
			log.Fatal(err)
		}

		dep.Licenses = string(licensePayload)

		if dep.CPE == "" {
			dep.CPE = fmt.Sprintf("cpe:2.3:a:tini_project:tini:%s:*:*:*:*:*:*:*", strings.TrimPrefix(dep.Version, "v"))
		}

		payload, err := json.Marshal(dep)
		if err != nil {
			log.Fatal(err)
		}

		var dispatch struct {
			EventType     string          `json:"event_type"`
			ClientPayload json.RawMessage `json:"client_payload"`
		}

		dispatch.EventType = "tini-test"
		dispatch.ClientPayload = json.RawMessage(payload)

		payloadData, err := json.Marshal(&dispatch)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(payloadData))
		// // send dispatch w all the info for each version
		// req, err := http.NewRequest("POST", "https://api.github.com/repos/paketo-buildpacks/dep-server/dispatches", bytes.NewBuffer(payloadData))
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))

		// resp2, err := http.DefaultClient.Do(req)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// // make sure we get a 200 status code
		// if resp2.StatusCode != http.StatusOK {
		// 	fmt.Println(resp2.StatusCode)
		// 	log.Fatal(err)
		// }
		// fmt.Printf("Success version %s!\n", dep.Version)
	}

	fmt.Println("Success!")

}
