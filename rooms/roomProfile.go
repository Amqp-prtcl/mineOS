package rooms

import (
	"encoding/json"
	"os"

	"github.com/Amqp-prtcl/snowflakes"
)

type RoomProfile struct {
	ID      snowflakes.ID `json:"id"`
	Name    string        `json:"name"`
	Emails  []string      `json:"emails"`
	JarPath string        `json:"jarpath"`
}

func LoadProfiles(file string) ([]RoomProfile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	var profiles = []RoomProfile{}
	err = json.NewDecoder(f).Decode(&profiles)
	f.Close()
	return profiles, err
}
