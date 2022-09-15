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
func Setup(defaults map[string]interface{}) error {
	var err error
	Config, err = config.LoadConfigFile(ConfigFile, config.Json, defaults)
	if err != nil {
		return err
	}
	return nil
}

func GetSecret() string //TODO

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
		fmt.Printf("[Globals] error in config system: %v\n", err)
	}
	return v
}

func (c ConfigTimeKey) WarnGet() time.Time {
	v, err := config.TimeKey(c).GetErr(Config)
	if err != nil {
		fmt.Printf("[Globals] error in config system: %v\n", err)
	}
	return v
}

const (
	DownloadFolder ConfigKey[string] = "download-folder"
	ProfilesFiles  ConfigKey[string] = "profiles-file"
	ServerFolder   ConfigKey[string] = "server-folder"
	UsersFile      ConfigKey[string] = "users-file"
	CacheFolder    ConfigKey[string] = "cache-folder"
	Time           ConfigTimeKey     = "epoch"
	AssetsFolder   ConfigKey[string] = "assets-folder"
	OfflineMode    ConfigKey[bool]   = "offline-mode"
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
