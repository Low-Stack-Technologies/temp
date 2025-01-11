package update

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"tech.low-stack.temp/cli/internal/env"
)

//go:embed version
var version string

type gitHubRelease struct {
	Name  string `json:"name"`
	Draft bool   `json:"draft"`
	Url   string `json:"html_url"`
}

func CheckVersion() {
	// Fetch the latest release from GitHub
	latestVersion, err := getLatestVersion()
	if err != nil || latestVersion == nil {
		fmt.Printf("Was unable to check for the latest version!")
		return
	}

	// Check if a new version has been released
	if strings.TrimSpace(latestVersion.Name) != strings.TrimSpace(version) {
		fmt.Println("A new version of the Temp CLI has been released!")
		fmt.Printf("Download at: %s\n", latestVersion.Url)
	}
}

func getLatestVersion() (*gitHubRelease, error) {
	res, err := http.Get(env.ReleasesUrl)
	if err != nil {
		return nil, err
	}

	releases := &[]gitHubRelease{}
	if err := json.NewDecoder(res.Body).Decode(releases); err != nil {
		return nil, err
	}

	for _, release := range *releases {
		if release.Draft {
			continue
		}

		return &release, nil
	}

	return nil, fmt.Errorf("no release found")
}
