package versions

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var (
	paperManifestUrl = "https://api.papermc.io/v2/projects/paper"
)

func (v paperVersion) GetVersionUrl() string {
	return fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s", v)
}

func (v paperVersion) paperGetBuildUrl(build int) string {
	return fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds/%v", v, build)
}

func (v paperVersion) paperDownloadUrl(build int, filename string) string {
	return fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds/%v/downloads/%s", v, build, filename)
}

type paperManifest struct {
	Versions []string `json:"versions"`
	mu       sync.RWMutex
}

type paperVersion string

func (m *paperManifest) GetType() ServerType { return Paper }
func (m paperVersion) GetType() ServerType   { return Paper }
func (m paperVersion) GetID() string         { return string(m) }

func paperGenerateManifest() (Manifest, error) {
	resp, err := http.Get(paperManifestUrl)
	if err != nil {
		return nil, err
	}
	var m = &paperManifest{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	resp.Body.Close()
	return m, err
}

func (m *paperManifest) GetVersionsList() []string {
	m.mu.RLock()
	var vrs = make([]string, len(m.Versions))
	copy(vrs, m.Versions)
	m.mu.RUnlock()
	return vrs
}

func (m *paperManifest) GetVersion(vrsID string) (Version, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, vrs := range m.Versions {
		if vrs == vrsID {
			return paperVersion(vrs), true
		}
	}
	return nil, false
}

func (v paperVersion) DownloadServer(path string) error {
	//TODO

	// get list of builds
	var builds = struct {
		Builds []int `json:"builds"`
	}{}
	err := retreiveStructFromUrl(v.GetVersionUrl(), &builds)
	if err != nil {
		return err
	}
	var build = builds.Builds[len(builds.Builds)-1]

	// from build control get hash and filename
	var control = struct {
		Downloads struct {
			Application struct {
				Name   string `json:"name"`
				Sha256 string `json:"sha256"`
			} `json:"application"`
		} `json:"downloads"`
	}{}
	err = retreiveStructFromUrl(v.paperGetBuildUrl(build), &control)
	if err != nil {
		return err
	}
	// download file
	return DownloadFile(path, v.paperDownloadUrl(build, control.Downloads.Application.Name), -1, control.Downloads.Application.Sha256, sha256.New)
}