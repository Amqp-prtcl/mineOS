package config

import (
	"encoding/json"
	"os"
	"reflect"
	"time"
)

var ( //defaults
	defaultConfig = config{
		AssetsFolder:        "",
		ServersFolder:       "",
		VersionsCacheFolder: "",
		UsersFile:           "",
		ServerProfilesFile:  "",
		BuildToolsFolder:    "",
		Epoch:               time.Unix(0, 0),
	}

	secret = "//TODO"

	Config *config
)

type config struct {
	AssetsFolder        string `json:"assets-folder"`
	ServersFolder       string `json:"servers-folder"`
	VersionsCacheFolder string `json:"versions-cache-folder"`

	Epoch time.Time `json:"epoch"`

	UsersFile          string `json:"users-file"`
	ServerProfilesFile string `json:"server-profiles-file"`

	BuildToolsFolder string `json:"build-tools-folder"`
}

func LoadConfig(path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	var c = new(config)
	err = json.NewDecoder(f).Decode(c)
	f.Close()
	if err != nil {
		return err
	}
	Config = c
	verifyConfig()
	return nil
}

func verifyConfig() {
	if Config == nil {
		*Config = defaultConfig
		return
	}
	v := reflect.ValueOf(Config)
	defaultV := reflect.ValueOf(defaultConfig)
	l := v.NumField()
	for i := 0; i < l; i++ {
		if v.Field(i).IsZero() {
			v.Field(i).Set(defaultV.Field(i))
		}
	}
}

func SaveConfig(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	if Config != nil {
		err = json.NewEncoder(f).Encode(Config)
	} else {
		err = json.NewEncoder(f).Encode(defaultConfig)
	}
	f.Close()
	return err
}

func GetSecret() string {
	return secret
}
