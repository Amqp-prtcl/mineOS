package versions

import (
	"encoding/json"
	"mineOS/globals"
	"os"
	"path/filepath"
	"sync"
)

func getCacheSrvTypeFolder(srvType ServerType) string {
	return filepath.Join(globals.CacheFolder.WarnGet(), string(srvType))
}

func getCacheVrsIDFolder(srvType ServerType, vrsID string) string {
	return filepath.Join(getCacheSrvTypeFolder(srvType), vrsID)
}

func getCacheVrsIDFile(srvType ServerType, vrsID string) string {
	return filepath.Join(getCacheVrsIDFolder(srvType, vrsID), "server.jar")
}

var (
	cache   = map[string]interface{}{}
	cachemu = sync.RWMutex{}
)

func loadCache(cachePath string) error {
	if cachePath == "" {
		cachePath = globals.CacheFolder.WarnGet()
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
		cachePath = globals.CacheFolder.WarnGet()
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

/*func cacheDelete(key string) {
	cachemu.Lock()
	defer cachemu.Unlock()
	delete(cache, key)
}*/
