package manager

import (
	"encoding/json"
	"fmt"
	"mineOS/rooms"
	"mineOS/servers"
	"mineOS/versions"
	"sync"

	"github.com/Amqp-prtcl/snowflakes"
	"github.com/gorilla/websocket"
)

var (
	M = NewManager()
)

type Manager struct {
	Rooms   []*rooms.Room
	roomsmu sync.RWMutex

	list   []*websocket.Conn
	listmu sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		Rooms:   []*rooms.Room{},
		roomsmu: sync.RWMutex{},
		list:    []*websocket.Conn{},
		listmu:  sync.Mutex{},
	}
}

// if file arg is empty, it is fetched from config file
func (m *Manager) LoadRooms(file string) error {
	m.roomsmu.Lock()
	defer m.roomsmu.Unlock()
	if len(m.Rooms) != 0 {
		//TODO
		return fmt.Errorf("double loading of rooms")
	}
	profiles, err := rooms.LoadProfiles(file)
	if err != nil {
		return err
	}
	for _, p := range profiles {
		m.Rooms = append(m.Rooms, rooms.NewRoom(p, m.OnStateChange))
	}
	m.roomsmu.Unlock()
	return nil
}

func (m *Manager) GetRoombyID(id snowflakes.ID) (*rooms.Room, bool) {
	m.roomsmu.RLock()
	defer m.roomsmu.RUnlock()

	for _, room := range m.Rooms {
		if room.Srv.ID == id {
			return room, true
		}
	}
	return nil, false
}

func (m *Manager) AddConn(conn *websocket.Conn) {
	m.listmu.Lock()
	m.list = append(m.list, conn)
	m.listmu.Unlock()
}

func (m *Manager) OnStateChange(srv *servers.Server) {
	if len(m.list) == 0 {
		return
	}

	data, _ := json.Marshal(&struct {
		ID    snowflakes.ID `json:"id"`
		State string        `json:"state"`
	}{})

	cn := []int{}
	m.listmu.Lock()
	for i := range m.list {
		err := m.list[i].WriteMessage(websocket.TextMessage, data)
		if err != nil {
			cn = append(cn, i)
		}
	}

	if len(cn) != 0 {
		for i := len(cn) - 1; i >= 0; i-- {
			m.list[cn[i]].Close()
			m.list[cn[i]] = nil
			m.list[cn[i]] = m.list[len(m.list)-1]
			m.list = m.list[0 : len(m.list)-1]
		}
	}
	m.listmu.Unlock()
}

func (m *Manager) NewRoom(prof *rooms.RoomProfile) bool {
	m.roomsmu.Lock()
	defer m.roomsmu.Unlock()
	for _, rm := range m.Rooms {
		if rm.Profile.ID == prof.ID || rm.Profile.Name == prof.Name {
			return false
		}
	}
	m.Rooms = append(m.Rooms, rooms.NewRoom(prof, m.OnStateChange))
	return true
}

func (m *Manager) MarshalServerList() []byte {
	type a struct {
		ID         snowflakes.ID       `json:"id"`
		Name       string              `json:"name"`
		ServerType versions.ServerType `json:"server-type"`
		VersionID  string              `json:"version-id"`
		State      servers.ServerState `json:"state"`
	}
	m.roomsmu.RLock()
	var srvs = make([]a, 0, len(m.Rooms))

	for _, r := range m.Rooms {
		srvs = append(srvs, a{
			ID:         r.Srv.ID,
			Name:       r.Profile.Name,
			ServerType: r.Profile.Type,
			VersionID:  r.Profile.VersionID,
			State:      r.Srv.State,
		})
	}
	m.roomsmu.RUnlock()
	data, _ := json.Marshal(srvs)
	return data
}
