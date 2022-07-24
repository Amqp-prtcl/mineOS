package versions

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
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
	Versions  []*vanillaVersion  `json:"versions"`
	cacheVers []*vanillaCache    `json:"-"`
	cacheDown []*vanillaDownload `json:"-"`
	mu        sync.RWMutex       `json:"-"`
}

type vanillaVersion struct {
	ID    string    `json:"id"`
	Type  string    `json:"type"`
	URL   string    `json:"url"`
	Time  time.Time `json:"time"`
	RTime time.Time `json:"releaseTime"`
}

type vanillaCache struct {
	ID   string
	Type string
	Sha1 string
	Path string
}

type vanillaDownload struct {
	vers  *vanillaVersion
	cache *vanillaCache
	m     *vanillaManifest
	done  bool
	cs    []chan *vanillaCache // is nil if success is false
	mu    sync.Mutex
}

func (d *vanillaDownload) waitFor() chan *vanillaCache {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.done {
		c := make(chan *vanillaCache, 1)
		c <- d.cache
		return c
	}
	c := make(chan *vanillaCache, 1)
	d.cs = append(d.cs, c)
	return c
}

// should be run as goroutine
//
// returned err value can be ignored and might only be used as debug or error message
func (d *vanillaDownload) download(path string) error {
	var meta = struct {
		Downloads struct {
			Server struct {
				Sha1 string `json:"sha1"`
				Size int64  `json:"size"`
				Url  string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}{}
	err := RetrieveStructFromUrl(d.vers.URL, &meta)
	if err != nil {
		return err
	}
	defer func(m *vanillaManifest, d *vanillaDownload) {
		go func(m *vanillaManifest, d *vanillaDownload) {
			m.mu.Lock()
			defer m.mu.Unlock()
			for i, down := range m.cacheDown {
				if down == d {
					m.cacheDown[i] = nil
					m.cacheDown[i] = m.cacheDown[len(m.cacheDown)]
					m.cacheDown = m.cacheDown[:len(m.cacheDown)-1]
					return
				}
			}
		}(d.m, d)
	}(d.m, d)
	err = DownloadFile(path, meta.Downloads.Server.Url, meta.Downloads.Server.Size, meta.Downloads.Server.Sha1, sha1.New)
	if err == nil {
		d.cache = &vanillaCache{
			ID:   d.vers.ID,
			Type: d.vers.Type,
			Sha1: meta.Downloads.Server.Sha1,
			Path: path,
		}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.done = true
	for _, c := range d.cs {
		c <- d.cache
	}
	return err
}

func (m *vanillaManifest) GetType() ServerType { return Vanilla }

//TODO sync mith cache
func vanillaGenerateManifest(offline bool) (Manifest, error) {
	var m = &vanillaManifest{}
	var err error
	if !offline {
		err = RetrieveStructFromUrl(vanillaManifestUrl, &m)
		/*for i := range m.Versions {
			m.Versions[i].m = m
		}*/
	}
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

func (v vanillaCache) downloadServer(path string) error {
	hash, err := GetSum(v.Path, sha1.New)
	if err != nil {
		return err
	}
	if hash != v.Sha1 {
		return fmt.Errorf("invalid checksum: cache control -> %q but cached file -> %q", v.Sha1, hash)
	}

	src, err := os.Open(v.Path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(src, dst)
	if err != nil {
		return err
	}

	hash, err = GetSum(path, sha1.New)
	if err != nil {
		return err
	}
	if hash != v.Sha1 {
		return fmt.Errorf("invalid checksum: cache control -> %q but server jar -> %q", v.Sha1, hash)
	}
	return nil
}

//TODO save cache before ??
func (m *vanillaManifest) SyncWithCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheVers = []*vanillaCache{}

	inter, ok := cacheGet(string(m.GetType()))
	if !ok {
		fmt.Printf("vanilla cache empty.\n")
		return
	}
	cache, ok := inter.([]interface{})
	if !ok {
		fmt.Printf("invalid vanilla cache.\n")
		return
	}

	for _, v := range cache {
		ver, ok := v.(map[string]interface{})
		if !ok {
			fmt.Printf("invalid version entry in vanilla cache.\n")
			continue
		}
		var a = &vanillaCache{}
		if i := DecodeMapToStruct(ver, &a); i != 4 {
			fmt.Printf("invalid version entry in vanilla cache: only found %v fields.\n", i)
			continue
		}
		m.cacheVers = append(m.cacheVers, a)
	}
}

// It is caller's responsibility to lock Mutexes
func (m *vanillaManifest) getCacheVersion(vrsID string) (*vanillaCache, bool)

// It is caller's responsibility to lock Mutexes
func (m *vanillaManifest) getDownloadingVersion(vrsID string) (*vanillaDownload, bool)

// It is caller's responsibility to lock Mutexes
func (m *vanillaManifest) getVersion(vrsID string) (*vanillaVersion, bool)

func (m *vanillaManifest) DownloadServer(vrsID string, path string) error {
	m.mu.Lock()
	if cache, ok := m.getCacheVersion(vrsID); ok {
		m.mu.Unlock()
		return cache.downloadServer(path)
	}

	if down, ok := m.getDownloadingVersion(vrsID); ok {
		m.mu.Unlock()
		cache := <-down.waitFor()
		if cache == nil {
			return fmt.Errorf("[vanilla manifest] download of versionID %v failed", vrsID)
		}
		return cache.downloadServer(path)
	}

	if vrs, ok := m.getVersion(vrsID); ok {
		var down = &vanillaDownload{
			vers: vrs,
			m:    m,
			cs:   []chan *vanillaCache{},
		}
		m.cacheDown = append(m.cacheDown, down)
		m.mu.Unlock()
		cache := <-down.waitFor()
		if cache == nil {
			return fmt.Errorf("[vanilla manifest] download of versionID %v failed", vrsID)
		}
		return cache.downloadServer(path)
	}

	m.mu.Unlock()
	return ErrVerIdNotFound
}
