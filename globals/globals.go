package globals

import (
	"fmt"
	"time"

	"github.com/Amqp-prtcl/config"
)

const ConfigFile string = "config.json"

var (
	Config *config.Config
)

// set to nil for no defaults
func Setup() error {
	var err error
	Config, err = config.LoadConfigFile(ConfigFile, config.Json)
	if err != nil {
		return err
	}
	return nil
}

func GetSecret() string {
	return "HUFE7843JIFE09083//fe="
} //TODO

func GetConfig(key string) (interface{}, bool) {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (GetConfig)\n")
		return nil, false
	}
	return Config.Get(key)
}
func GetConfigG[T config.Getable](key string) (T, bool) {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (GetConfigG)\n")
		var t T
		return t, false
	}
	return config.Get[T](Config, key)
}
func PutConfig(key string, v interface{}) {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (PutConfig)\n")
		return
	}
	Config.Put(key, v)
}

func SaveConfig() error {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (PutConfig)\n")
		return nil
	}
	return Config.SaveFile()
}

//Config Keys

type ConfigKey[T config.Keyable] config.Key[T]
type ConfigTimeKey config.TimeKey

func (c ConfigKey[T]) WarnGet() T {
	v, err := config.Key[T](c).GetErr(Config)
	if err != nil {
		fmt.Printf("[Globals] error in config system (key: %v): %v\n", c, err)
	}
	return v
}

func (c ConfigTimeKey) WarnGet() time.Time {
	v, err := config.TimeKey(c).GetErr(Config)
	if err != nil {
		fmt.Printf("[Globals] error in config system (key: %v): %v\n", c, err)
	}
	return v
}

var (
	DownloadFolder = ConfigKey[string]{"download-folder", "/Users/temp/MineOs/downloads/"}
	ProfilesFiles  = ConfigKey[string]{"profiles-file", "/Users/temp/MineOs/profiles.json"}
	ServerFolder   = ConfigKey[string]{"server-folder", "/Users/temp/MineOs/servers/"}
	UsersFile      = ConfigKey[string]{"users-file", "/Users/temp/MineOs/users.json"}
	CacheFolder    = ConfigKey[string]{"cache-folder", "/Users/temp/MineOs/cache/"}
	Time           = ConfigTimeKey{"epoch", time.Now()}
	AssetsFolder   = ConfigKey[string]{"assets-folder", "/Users/temp/MineOs/assets/"}
	OfflineMode    = ConfigKey[bool]{"offline-mode", false}
)

type MultiError []error

func (m MultiError) Error() string {
	var str string = "multiple errors: "
	for i, e := range m {
		str += fmt.Sprintf("[%v] %v. ", i, e)
	}
	return str
}

func (m *MultiError) Append(e error) {
	if e == nil {
		return
	}
	*m = append(*m, e)
}

// IsEmpty return true if all errors in m are nil
func (m MultiError) IsEmpty() bool {
	return len(m) == 0
}

// ToErr return nil if IsEmpty return true, otherwise ToErr return m
//
// It is used to ease chaining of operations
func (m *MultiError) ToErr() error {
	if m.IsEmpty() {
		return nil
	}
	return m
}
