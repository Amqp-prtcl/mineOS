package rooms

import (
	"encoding/json"
	"mineOS/config"
	"mineOS/versions"
	"os"

	"github.com/Amqp-prtcl/snowflakes"
)

type RoomProfile struct {
	ID        snowflakes.ID       `json:"id"`
	Type      versions.ServerType `json:"server-type"`
	VersionID string              `json:"version-id"`
	Name      string              `json:"name"`
	Emails    []string            `json:"emails"`
	JarPath   string              `json:"jarpath"`
}

// if file arg if empty, it will be fetch from config file
func LoadProfiles(file string) ([]*RoomProfile, error) {
	if file == "" {
		file = config.Config.ServersFolder
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	var profiles = []*RoomProfile{}
	err = json.NewDecoder(f).Decode(&profiles)
	f.Close()
	return profiles, err
}
