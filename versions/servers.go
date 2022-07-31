package versions

type ServerType string

const (
	Vanilla ServerType = "VANILLA"
	//Paper   ServerType = "PAPERMC"
)

func GetServerTypes() []ServerType {
	return []ServerType{Vanilla /*Paper*/}
}

type Manifest interface {
	GetVersionsList() []string
	//GetVersion(vrsid string) (Version, bool)

	// if vrsID is invalid, DownloadServer must respond with ErrVerIdNotFound
	DownloadServer(vrsID string, path string) error
	GetType() ServerType
}

var (
	vanillaM Manifest
	//paperM   Manifest
)

func Setup(cachePath string, offline bool) error {
	err := loadCache(cachePath)
	if err != nil {
		return err
	}

	vanillaM, err = vanillaGenerateManifest(offline)
	if err != nil {
		return err
	}
	//paperM, err = paperGenerateManifest(config.Config.OfflineMode)
	return err
}

func Save(cachePath string) error {
	return saveCache(cachePath)
}

func GetManifestByServerType(srvType ServerType) (Manifest, bool) {
	switch srvType {
	case Vanilla:
		return vanillaM, true
	/*case Paper:
	return paperM, true*/
	default:
		return nil, false
	}
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
