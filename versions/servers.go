package versions

import (
	"mineOS/globals"
	"strings"
)

type ServerType string

const (
	Vanilla ServerType = "VANILLA"
	//Paper   ServerType = "PAPERMC"
)

func ToServerType(str string) ServerType {
	return ServerType(strings.ToUpper(str))
}

func GetServerTypes() []ServerType {
	return []ServerType{Vanilla /*,Paper*/}
}

func ForEachSrvType(f func(srvType ServerType)) {
	for _, v := range GetServerTypes() {
		f(v)
	}
}

func ForEachManifest(f func(m Manifest)) {
	for _, v := range manifests {
		f(v)
	}
}

var (
	manifests = []Manifest{}
)

func Setup(cachePath string, offline bool) error {
	err := loadCache(cachePath)
	if err != nil {
		return err
	}

	m, err := vanillaGenerateManifest(offline)
	if err != nil {
		return err
	}
	manifests = append(manifests, m)
	//paperM, err = paperGenerateManifest(config.Config.OfflineMode)
	return err
}

type Manifest interface {
	GetVersionsList() []string

	// if vrsID is invalid, DownloadServer must respond with ErrVerIdNotFound
	DownloadServer(vrsID string, path string) error
	GetType() ServerType

	// if vrsID does not exists, ClearCache should return nil
	ClearCache(vrsID string) error

	// error should be of type globals.MultiError
	ClearCacheAll() error
}

func SaveCache(cachePath string) error {
	return saveCache(cachePath)
}

func GetManifestByServerType(srvType ServerType) (Manifest, bool) {
	for _, m := range manifests {
		if m.GetType() == srvType {
			return m, true
		}
	}
	return nil, false
}

func GetVersionIdsBuServerType(srvType ServerType) ([]string, bool) {
	m, ok := GetManifestByServerType(srvType)
	if !ok {
		return nil, false
	}
	return m.GetVersionsList(), true
}

func DownloadServerByServerType(srvType ServerType, vrsID string, path string) error {
	m, ok := GetManifestByServerType(srvType)
	if !ok {
		return ErrSrvTypeNotFound
	}
	return m.DownloadServer(vrsID, path)
}

// ClearCache clears all versions it can for all versions
// (if there is an error, it will be of type globals.MultiError)
//
// To clear only part of the cache see Manifest.ClearCache or Manifest.ClearCacheAll
func ClearCacheAll() error {
	var e = globals.MultiError{}
	ForEachManifest(func(m Manifest) {
		e.Append(m.ClearCacheAll())
	})
	return e.ToErr()
}
