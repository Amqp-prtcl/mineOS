package papermc

import (
	"encoding/json"
	"mineOS/config"
	"net/http"
	"sync"
)

var (
	M *Manifest
)

func Setup() error {
	var err error
	M, err = loadVersion("")
	return err
}

type Version string

func (v Version) String() string {
	return string(v)
}

type Manifest struct {
	Versions []Version `json:"versions"`
	mu       sync.RWMutex
}

func loadVersion(manifestUrl string) (*Manifest, error) {
	if manifestUrl == "" {
		manifestUrl = config.Config.PaperMcManifestURL
	}
	resp, err := http.Get(manifestUrl)
	if err != nil {
		return nil, err
	}
	var m = &Manifest{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	resp.Body.Close()
	return m, err
}

func (m *Manifest) GetVersionList() []string {
	var vrs = []string{}
	m.mu.RLock()
	for _, v := range m.Versions {
		vrs = append(vrs, v.String())
	}
	m.mu.RUnlock()
	return vrs
}

func (m *Manifest) GetVersion(id string) (*Version, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, vrs := range m.Versions {
		if vrs.String() == id {
			return &vrs, true
		}
	}
	return nil, false
}
