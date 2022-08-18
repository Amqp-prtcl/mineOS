package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"mineOS/globals"
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
		file = globals.ProfilesFiles.WarnGet()
	}
	var profiles = []*RoomProfile{}

	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil
		}
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&profiles)
	fmt.Printf("loaded %v server profiles\n", len(profiles))
	return profiles, err
}

func SaveProfiles(file string, l []*RoomProfile) error {
	if file == "" {
		file = globals.ProfilesFiles.WarnGet()
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(l)
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
	var serverDir = filepath.Join(globals.ServerFolder.WarnGet(), profile.ID.String())
	err := os.MkdirAll(serverDir, 0666)
	if err != nil {
		return nil, err
	}
	profile.JarPath = filepath.Join(serverDir, "server.jar")

	var ok = false
	defer func(ok *bool) {
		if !(*ok) {
			go os.RemoveAll(serverDir)
		}
	}(&ok)

	// 2. download jar file (differs from serverType)
	err = versions.DownloadServerByServerType(profile.Type, profile.VersionID, profile.JarPath)
	if err != nil {
		if err == versions.ErrVerIdNotFound || err == versions.ErrSrvTypeNotFound {
			return nil, fmt.Errorf("unknown minecraft version id: %v (server type: %v)", profile.VersionID, profile.Type)
		}
		return nil, err
	}

	// 3. agree to eula by running jar once and editing 'eula.txt'
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
	ok = true
	return profile, nil
}
