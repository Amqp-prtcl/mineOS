package main

import (
	"fmt"
	"mineOS/servers"
	"sync"

	"github.com/Amqp-prtcl/snowflakes"
	"github.com/gorilla/websocket"
)

type Room struct {
	Srv *servers.Server

	conns []*websocket.Conn
	mu    sync.Mutex

	cmds chan string
}

func NewRoom(jarPath string, name string, id snowflakes.ID) *Room {
	r := &Room{
		Srv:   servers.NewServer(jarPath, name, id),
		conns: []*websocket.Conn{},
		mu:    sync.Mutex{},
		cmds:  make(chan string, 1),
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
}
