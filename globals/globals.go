package globals

import (
	"fmt"
	"mineOS/config"
)

const ConfigFile string = "config.json"

var (
	Config *config.Config
)

// set to nil for no defaults
func Setup(defaults map[string]interface{}) error {
	var err error
	Config, err = config.LoadConfigFile(ConfigFile, defaults)
	if err != nil {
		return err
	}
	return nil
}

func GetSecret() string //TODO

func GetConfig(key string) (interface{}, bool) {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (GetConfig)")
		return nil, false
	}
	return Config.Get(key)
}
func GetConfigG[T any](key string) (T, bool) {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (GetConfigG)")
		var t T
		return t, false
	}
	return config.Get[T](Config, key)
}
func PutConfig(key string, v interface{}) {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (PutConfig)")
		return
	}
	Config.Put(key, v)
}

func SaveConfig() error {
	if Config == nil {
		fmt.Printf("[Globals] Access to nil config (PutConfig)")
		return nil
	}
	return Config.SaveFile()
}

func WarnConfigGet[T any](key string) T {
	t, ok := GetConfigG[T](key)
	if !ok {
		fmt.Printf("[Globals] error in config system")
	}
	return t
}
