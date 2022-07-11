package versions

import (
	"crypto/sha1"
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
func (m *vanillaVersion) GetID() string        { return m.ID }

func vanillaGenerateManifest() (Manifest, error) {
	var m = &vanillaManifest{}
	return m, retreiveStructFromUrl(vanillaManifestUrl, &m)
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

func (v *vanillaVersion) DownloadServer(path string) error {
	var meta = struct {
		Downloads struct {
			Server struct {
				Sha1 string `json:"sha1"`
				Size int64  `json:"size"`
				Url  string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}{}
	err := retreiveStructFromUrl(v.URL, &meta)
	if err != nil {
		return err
	}
	return DownloadFile(path, meta.Downloads.Server.Url, meta.Downloads.Server.Size, meta.Downloads.Server.Sha1, sha1.New)
}