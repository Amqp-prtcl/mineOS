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

func setupCache() error {
	err := os.MkdirAll(config.Config.VersionsCacheFolder, 0664)
	if err != nil {
		return err
	}
	f, err := os.Open(filepath.Join(config.Config.VersionsCacheFolder, "versions.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return saveCache()
		}
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&cache)
}

func saveCache() error {
	f, err := os.Create(filepath.Join(config.Config.VersionsCacheFolder, "versions.json"))
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
