package rooms

import (
	"fmt"
	"mineOS/emails"
	"mineOS/servers"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	Srv     *servers.Server
	Profile *RoomProfile
	conns   []*websocket.Conn
	mu      sync.Mutex

	stateCallback func(*servers.Server)

	cmds chan string
}

func NewRoom(profile *RoomProfile, stateCallback func(*servers.Server)) *Room {
	r := &Room{
		Srv:           servers.NewServer(profile.JarPath, profile.ID),
		Profile:       profile,
		conns:         []*websocket.Conn{},
		mu:            sync.Mutex{},
		stateCallback: stateCallback,
		cmds:          make(chan string, 1),
	}

	r.Srv.OnLog = r.onLog
	r.Srv.OnStateChange = r.onStateChange
	return r
}

func (r *Room) Start() error {
	err := r.Srv.Start()
	if err != nil {
		return err
	}
	go r.cmdHandler()
	return nil
}

func (r *Room) Stop() {
	r.Srv.Stop()
}

func (r *Room) SendCommand(cmd string) {
	if r.Srv.State != servers.Closed {
		r.cmds <- cmd
	}
}

func (r *Room) cmdHandler() {
	for r.Srv.State != servers.Closed {
		select {
		case cmd := <-r.cmds:
			r.Srv.SendCommand(cmd)
		default:
			continue
		}
	}
}

func (r *Room) AddConn(conn *websocket.Conn) {
	r.mu.Lock()
	r.conns = append(r.conns, conn)
	r.mu.Unlock()
	go func(c *websocket.Conn, ch chan string) {
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			ch <- string(data)
		}
	}(conn, r.cmds)
}

func (r *Room) onLog(_ *servers.Server, log string) {
	cn := []int{}
	r.mu.Lock()
	for i := range r.conns {
		err := r.conns[i].WriteMessage(websocket.TextMessage, []byte(log))
		if err != nil {
			fmt.Println(err)
			cn = append(cn, i)
		}
	}

	if len(cn) != 0 {
		for i := len(cn) - 1; i >= 0; i-- {
			r.conns[cn[i]].Close()
			r.conns[cn[i]] = nil
			r.conns[cn[i]] = r.conns[len(r.conns)-1]
			r.conns = r.conns[0 : len(r.conns)-1]
		}
	}
	r.mu.Unlock()
}

func (r *Room) onStateChange(_ *servers.Server) {
	cn := []int{}
	r.mu.Lock()
	for i := range r.conns {
		err := r.conns[i].WriteMessage(websocket.TextMessage, []byte(r.Srv.State))
		if err != nil {
			cn = append(cn, i)
		}
	}

	if len(cn) != 0 {
		for i := len(cn) - 1; i >= 0; i-- {
			r.conns[cn[i]].Close()
			r.conns[cn[i]] = nil
			r.conns[cn[i]] = r.conns[len(r.conns)-1]
			r.conns = r.conns[0 : len(r.conns)-1]
		}
	}
	r.mu.Unlock()
	if r.stateCallback != nil {
		r.stateCallback(r.Srv)
	}
	switch r.Srv.State {
	case servers.Running:
		err := r.sendRunningEmail()
		if err != nil {
			fmt.Printf("Error sending running email(s): %v", err)
		}
	case servers.Closed:
		err := r.sendCloseMail()
		if err != nil {
			fmt.Printf("Error sending closing email(s): %v", err)
		}
	}
}

func (r *Room) sendRunningEmail() error {
	var subject = fmt.Sprintf("MineOS: Server %s (id: %s) Running.", r.Profile.Name, r.Profile.ID.String())
	var body = fmt.Sprintf("Server %s (id: %s) is now running if this is unintentional or unexpected please log in in order to resolve possible issue.", r.Profile.Name, r.Profile.ID.String())
	return emails.SendEmail(r.Profile.Emails, subject, body)
}

func (r *Room) sendCloseMail() error {
	var subject = fmt.Sprintf("MineOS: Server %s (id: %s) Closed.", r.Profile.Name, r.Profile.ID.String())
	var body = fmt.Sprintf("Server %s (id: %s) has closed if this is unintentional or unexpected please log in in order to resolve possible issue.", r.Profile.Name, r.Profile.ID.String())
	return emails.SendEmail(r.Profile.Emails, subject, body)
}
