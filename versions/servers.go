package versions

type ServerType string

const (
	Vanilla ServerType = "VANILLA"
	Paper   ServerType = "PAPERMC"
)

func GetServerTypes() []ServerType {
	return []ServerType{Vanilla, Paper}
}

type Manifest interface {
	GetVersionsList() []string
	GetVersion(vrsid string) (Version, bool)
	GetType() ServerType
}

type Version interface {
	GetID() string
	DownloadServer(filepath string) error
	GetType() ServerType
}

var (
	vanillaM Manifest
	paperM   Manifest
)

func Setup() error {
	var err error
	vanillaM, err = vanillaGenerateManifest()
	if err != nil {
		return err
	}
	paperM, err = paperGenerateManifest()
	return err
}

func GetManifestByServerType(srvType ServerType) (Manifest, bool) {
	switch srvType {
	case Vanilla:
		return vanillaM, true
	case Paper:
		return paperM, true
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

func GetVersionByServerTypeAndVersionId(srvType ServerType, vrsID string) (Version, bool) {
	m, ok := GetManifestByServerType(srvType)
	if !ok {
		return nil, false
	}
	return m.GetVersion(vrsID)
}
