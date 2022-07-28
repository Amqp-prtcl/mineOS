package versions

import (
	"encoding/json"
	"mineOS/config"
	"os"
	"path/filepath"
	"sync"
)

var (
	cache   = map[string]interface{}{}
	cachemu = sync.RWMutex{}
)

func loadCache(cachePath string) error {
	if cachePath == "" {
		cachePath = config.Config.VersionsCacheFolder
	}
	err := os.MkdirAll(cachePath, 0666)
	if err != nil {
		return err
	}
	f, err := os.Open(filepath.Join(cachePath, "versions.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return saveCache(cachePath)
		}
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&cache)
}

func saveCache(cachePath string) error {
	if cachePath == "" {
		cachePath = config.Config.VersionsCacheFolder
	}
	f, err := os.Create(filepath.Join(cachePath, "versions.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(cache)
}

func cacheGet(key string) (interface{}, bool) {
	cachemu.RLock()
	defer cachemu.RUnlock()
	a, b := cache[key]
	return a, b
}

func cachePut(key string, val interface{}) {
	cachemu.Lock()
	defer cachemu.Unlock()
	cache[key] = val
}
