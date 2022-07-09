package versions

type ServerType string

const (
	Vanilla ServerType = "VANILLA"
	Bukkit  ServerType = "BUKKIT"
	Spigot  ServerType = "SPIGOT"
	PaperMc ServerType = "PAPERMC"
)

type Manifest interface {
	GetVersionsList() []string
	GetVersion(string) (Version, bool)
	GetType() ServerType
}

type Version interface {
	DownloadServer(string) error
	GetType() ServerType
}

var (
	vanillaM Manifest
)

func Setup() error {
	var err error
	vanillaM, err = VanillaGenerateManifest()
	if err != nil {
		return err
	}
	return nil
}

func GetManifestByServerType(srvType ServerType) (Manifest, bool) {
	switch srvType {
	case Vanilla:
		return vanillaM, true
	case Bukkit:
		//TODO
	case Spigot:
		//TODO
	case PaperMc:
		//TODO
	default:
		return nil, false
	}
	return nil, false
}

func GetVersionByServerTypeAndVersionId(srvType ServerType, vrsID string) (Version, bool) {
	m, ok := GetManifestByServerType(srvType)
	if !ok {
		return nil, false
	}
	return m.GetVersion(vrsID)
}
