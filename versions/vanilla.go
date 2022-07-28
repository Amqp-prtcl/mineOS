package versions

import (
	"crypto/sha1"
	"fmt"
	"io"
	"mineOS/config"
	"os"
	"path/filepath"
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
	mu        sync.Mutex         `json:"-"`
}

func (m *vanillaManifest) GetType() ServerType { return Vanilla }

//TODO sync mith cache
func vanillaGenerateManifest(offline bool) (Manifest, error) {
	var m = &vanillaManifest{}
	var err error
	if !offline {
		err = RetrieveStructFromUrl(vanillaManifestUrl, &m)
		if err != nil {
			return nil, err
		}
	}
	err = m.loadCache()
	return m, err
}

func (m *vanillaManifest) GetVersionsList() []string {
	m.mu.Lock()
	var vrs = make([]string, len(m.Versions))
	for i := range m.Versions {
		vrs[i] = m.Versions[i].ID
	}
	m.mu.Unlock()
	return vrs
}

type vanillaVersion struct {
	ID    string    `json:"id"`
	Type  string    `json:"type"`
	URL   string    `json:"url"`
	Time  time.Time `json:"time"`
	RTime time.Time `json:"releaseTime"`
}

type vanillaCache struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Sha1 string `json:"sha1"`
	Path string `json:"path"`
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
func (d *vanillaDownload) download() error {
	var path = filepath.Join(config.Config.VersionsCacheFolder, string(Vanilla), d.vers.ID, "server.jar")
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
					m.cacheDown[i] = m.cacheDown[len(m.cacheDown)-1]
					m.cacheDown = m.cacheDown[:len(m.cacheDown)-1]
					if d.cache != nil {
						m.cacheVers = append(m.cacheVers, d.cache)
					}
					m.saveCache()
					return
				}
			}
		}(d.m, d)
	}(d.m, d)
	err = DownloadFile(path, meta.Downloads.Server.Url, meta.Downloads.Server.Size, meta.Downloads.Server.Sha1, sha1.New)
	if err == nil {
		fmt.Println(3, err)
		d.cache = &vanillaCache{
			ID:   d.vers.ID,
			Type: d.vers.Type,
			Sha1: meta.Downloads.Server.Sha1,
			Path: path,
		}
	}
	d.mu.Lock()
	d.done = true
	for _, c := range d.cs {
		c <- d.cache
	}
	d.mu.Unlock()
	return err
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
		src.Close()
		return err
	}

	_, err = io.Copy(dst, src)
	src.Close()
	dst.Close()
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

// Locks mutexes
func (m *vanillaManifest) loadCache() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cacheVers = []*vanillaCache{}

	inter, ok := cacheGet(string(m.GetType()))
	if !ok {
		// just no cache present
		return nil
	}
	cache, ok := inter.([]interface{})
	if !ok {
		return fmt.Errorf("cannot load vanilla cache: failed to cast to []interface{} (got type: %T)", inter)
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
	return nil
}

//does NOT lock mutexes
func (m *vanillaManifest) saveCache() {
	cachePut(string(m.GetType()), m.cacheVers) // there may be race condition
}

// It is caller's responsibility to lock Mutexes
func (m *vanillaManifest) getCacheVersion(vrsID string) (*vanillaCache, bool) {
	for _, v := range m.cacheVers {
		if v.ID == vrsID {
			return v, true
		}
	}
	return nil, false
}

// It is caller's responsibility to lock Mutexes
func (m *vanillaManifest) getDownloadingVersion(vrsID string) (*vanillaDownload, bool) {
	for _, v := range m.cacheDown {
		if v.vers.ID == vrsID {
			return v, true
		}
	}
	return nil, false
}

// It is caller's responsibility to lock Mutexes
func (m *vanillaManifest) getVersion(vrsID string) (*vanillaVersion, bool) {
	for _, v := range m.Versions {
		if v.ID == vrsID {
			return v, true
		}
	}
	return nil, false
}

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
		go func() {
			err := down.download()
			if err != nil {
				fmt.Println(1, err)
			}
		}()
		cache := <-down.waitFor()
		if cache == nil {
			return fmt.Errorf("[vanilla manifest] download of versionID %v failed", vrsID)
		}
		return cache.downloadServer(path)
	}

	m.mu.Unlock()
	return ErrVerIdNotFound
}
