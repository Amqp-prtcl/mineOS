package config

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"time"
)

var ( //defaults
	defaultConfig = config{
		AssetsFolder:        "/Users/temp/Desktop/tasker/test/assets/",
		ServersFolder:       "/Users/temp/Desktop/tasker/test/servers/",
		VersionsCacheFolder: "/Users/temp/Desktop/tasker/test/versions/",
		DownloadFolder:      "/Users/temp/Desktop/tasker/test/downloads/",
		Epoch:               time.Unix(0, 0),
		OfflineMode:         false,
		UsersFile:           "/Users/temp/Desktop/tasker/test/users.json",
		ServerProfilesFile:  "/Users/temp/Desktop/tasker/test/servers.json",
		BuildToolsFolder:    "",
	}

	secret = "//TODO"

	Config *config
)

type config struct {
	AssetsFolder        string `json:"assets-folder"`
	ServersFolder       string `json:"servers-folder"`
	VersionsCacheFolder string `json:"versions-cache-folder"`
	DownloadFolder      string `json:"download-folder"`

	Epoch       time.Time `json:"epoch"`
	OfflineMode bool      `json:"offline-mode"`

	UsersFile          string `json:"users-file"`
	ServerProfilesFile string `json:"server-profiles-file"`

	BuildToolsFolder string `json:"build-tools-folder"`
}

func LoadConfig(path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	var c = new(config)
	err = json.NewDecoder(f).Decode(c)
	f.Close()
	if err != nil {
		if err == io.EOF {
			verifyConfig()
			return nil
		}
		return err
	}
	Config = c
	verifyConfig()
	return SaveConfig(path)
}

func verifyConfig() {
	if Config == nil {
		Config = new(config)
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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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
