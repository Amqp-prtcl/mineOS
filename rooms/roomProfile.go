package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"mineOS/config"
	"mineOS/versions"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Amqp-prtcl/snowflakes"
)

var (
	ServersNode = snowflakes.NewNode(0)
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

func GenerateRoom(name string, serverType versions.ServerType, versionID string) (*RoomProfile, error) {
	// protocol:
	// 1. generate id and create directory
	// 2. download jar file (differs from serverType)
	// 3. agree to eula by running jar once and editing 'eula.txt'

	var profile = &RoomProfile{
		Type:      serverType,
		VersionID: versionID,
		Name:      name,
	}
	// 1. generate id and create directory
	profile.ID = ServersNode.NewID()
	var serverDir = filepath.Join(config.Config.ServersFolder, profile.ID.String())
	err := os.MkdirAll(serverDir, 0777)
	if err != nil {
		return nil, err
	}
	profile.JarPath = filepath.Join(serverDir, "server.jar")

	// 2. download jar file (differs from serverType)
	vrs, ok := versions.GetVersionByServerTypeAndVersionId(profile.Type, profile.VersionID)
	if !ok {
		return nil, fmt.Errorf("unknown minecraft version id: %v (server type: %v)", profile.VersionID, profile.Type)
	}
	err = vrs.DownloadServer(profile.JarPath)
	if err != nil {
		return nil, err
	}

	// 3. agree to eula by running jar once and editing 'eula.txt'
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	cmd := exec.CommandContext(ctx, "java", "-jar", profile.JarPath, "nogui")
	cmd.Dir = serverDir
	err = cmd.Run()
	cancel()
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(serverDir, "eula.txt"), os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	err = f.Truncate(info.Size() - 6)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, 2)
	if err != nil {
		return nil, err
	}
	_, err = f.Write([]byte("true\n"))
	if err != nil {
		return nil, err
	}
	return profile, nil
}
