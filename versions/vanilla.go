package versions

import (
	"crypto/sha1"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	vanillaManifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
)

type vanillaManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []*vanillaVersion `json:"versions"`
	mu       sync.RWMutex
}

type vanillaVersion struct {
	ID    string    `json:"id"`
	Type  string    `json:"type"`
	URL   string    `json:"url"`
	Time  time.Time `json:"time"`
	RTime time.Time `json:"releaseTime"`
}

func (m *vanillaManifest) GetType() ServerType { return Vanilla }
func (m *vanillaVersion) GetType() ServerType  { return Vanilla }

func VanillaGenerateManifest() (Manifest, error) {
	resp, err := http.Get(vanillaManifestUrl)
	if err != nil {
		return nil, err
	}
	var m = &vanillaManifest{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	resp.Body.Close()
	return m, err
}

func (m *vanillaManifest) GetVersionsList() []string {
	m.mu.RLock()
	var vrs = make([]string, len(m.Versions))
	for i := range m.Versions {
		vrs[i] = m.Versions[i].ID
	}
	m.mu.RUnlock()
	return vrs
}

func (m *vanillaManifest) GetVersion(vrsID string) (Version, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, vrs := range m.Versions {
		if vrs.ID == vrsID {
			return vrs, true
		}
	}
	return nil, false
}

type vanillaVersionMeta struct {
	Downloads struct {
		Server vanillaFile `json:"server"`
	} `json:"downloads"`
}

type vanillaFile struct {
	Sha1 string `json:"sha1"`
	Size int64  `json:"size"`
	Url  string `json:"url"`
}

func (v *vanillaVersion) DownloadServer(path string) error {
	mf, err := getDownloadFileFromVersionUrl(v.URL)
	if err != nil {
		return err
	}
	return DownloadFile(path, mf.Url, mf.Size, mf.Sha1, sha1.New)
}

func getDownloadFileFromVersionUrl(vrsUrl string) (vanillaFile, error) {
	var meta = vanillaVersionMeta{}
	resp, err := http.Get(vrsUrl)
	if err != nil {
		return meta.Downloads.Server, err
	}
	err = json.NewDecoder(resp.Body).Decode(&meta)
	return meta.Downloads.Server, err
}
