package vanilla

import (
	"encoding/json"
	"mineOS/config"
	"net/http"
	"sync"
	"time"
)

type Manifest struct {
	L  Latest     `json:"latest"`
	Vs []*Version `json:"versions"`
	mu sync.Mutex
}

type Latest struct {
	R string `json:"release"`
	S string `json:"snapshot"`
}

type Version struct {
	ID    string    `json:"id"`
	T     string    `json:"type"`
	URL   string    `json:"url"`
	Time  time.Time `json:"time"`
	RTime time.Time `json:"releaseTime"`
}

// if manifest url is empty LoadVersion will automatically fetch the url from config file
func LoadVersions(manifestUrl string) (*Manifest, error) {
	if manifestUrl == "" {
		manifestUrl = config.Config.VanillaManifestURL
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
	m.mu.Lock()
	var vrs = make([]string, len(m.Vs))
	for i := range m.Vs {
		vrs[i] = m.Vs[i].ID
	}
	m.mu.Unlock()
	return vrs
}

func (m *Manifest) GetVersion(id string) (*Version, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, vrs := range m.Vs {
		if vrs.ID == id {
			return vrs, true
		}
	}
	return nil, false
}
