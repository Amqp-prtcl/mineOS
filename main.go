package main

import (
	"fmt"
	"log"
	"mineOS/servers"
	"net/http"
	"strings"
	"sync"

	"github.com/Amqp-prtcl/snowflakes"
	"github.com/gorilla/websocket"
)

const (
	jarPath               = "C:/Users/Luca/Desktop/test server 1.19/spigot-1.19.jar"
	id      snowflakes.ID = "1234"
)

var (
	server  *servers.Server
	manager = &Manager{
		conns: []*websocket.Conn{},
		mu:    sync.RWMutex{},
		input: make(chan string, 10),
	}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 20,
	WriteBufferSize: 1024 * 20,
}

func main() {
	go manager.listen()

	server = servers.NewServer(jarPath, "tg6", id)
	server.OnLog = manager.onLog
	server.OnStateChange = manager.onState

	http.HandleFunc("/", Handler)
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		panic(err)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.TrimSuffix(path, "/")

	if r.Method == http.MethodGet && path == "/favicon.ico" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet && path == "" {
		http.Redirect(w, r, "/servers", http.StatusPermanentRedirect)
		return
	}

	fmt.Println(r.Method)
	fmt.Println(path)

	if r.Method == http.MethodGet && path == "/servers" {
		w.Write([]byte("<!DOCTYPE html><html><head><meta charset='utf-8'><title>Servers</title></head><body><ul><a href=/servers/" + server.ID.String() + ">Server: " + server.Name + "</a> state: " + string(server.State) + "</ul></body></html>"))
		return
	}

	if r.Method == http.MethodGet && strings.HasSuffix(path, "t.js") {
		http.ServeFile(w, r, "t.js")
		return
	}

	if r.Method == http.MethodPost && strings.HasSuffix(path, "start") {
		server.Start()
		return
	}

	if r.Method == http.MethodPost && strings.HasSuffix(path, "stop") {
		server.Stop()
		return
	}

	if strings.HasSuffix(path, "/ws") {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		manager.addConn(conn)
		return
	}

	if r.Method == http.MethodGet && strings.HasPrefix(path, "/servers/") {
		http.ServeFile(w, r, "t.html")
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

type Manager struct {
	conns []*websocket.Conn
	mu    sync.RWMutex

	input chan string
}

func (m *Manager) listen() {
	for {
		str := <-m.input
		if server.State == servers.Running {
			server.SendCommand(str)
		}

	}
}

func (m *Manager) addConn(c *websocket.Conn) {
	m.mu.Lock()
	m.conns = append(m.conns, c)
	m.mu.Unlock()
	go listen(c, m.input)
}

func listen(c *websocket.Conn, ch chan string) {
	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			return
		}
		ch <- string(data)
	}
}

func (m *Manager) onLog(s *servers.Server, log string) {
	cn := []int{}
	m.mu.RLock()
	for i := range m.conns {
		err := m.conns[i].WriteMessage(websocket.TextMessage, []byte(log))
		if err != nil {
			fmt.Println(err)
			cn = append(cn, i)
		}
	}
	m.mu.RUnlock()

	if len(cn) != 0 {
		m.mu.Lock()
		for i := len(cn) - 1; i >= 0; i-- {
			m.conns[cn[i]].Close()
			m.conns[cn[i]] = nil
			m.conns[cn[i]] = m.conns[len(m.conns)-1]
			m.conns = m.conns[0 : len(m.conns)-1]
		}
		m.mu.Unlock()
	}
}

func (m *Manager) onState(s *servers.Server) {
	cn := []int{}
	m.mu.RLock()
	for i := range m.conns {
		err := m.conns[i].WriteMessage(websocket.TextMessage, []byte("$"+s.State))
		if err != nil {
			cn = append(cn, i)
		}
	}
	m.mu.RUnlock()

	if len(cn) != 0 {
		m.mu.Lock()
		for i := len(cn) - 1; i >= 0; i-- {
			m.conns[cn[i]].Close()
			m.conns[cn[i]] = nil
			m.conns[cn[i]] = m.conns[len(m.conns)-1]
			m.conns = m.conns[0 : len(m.conns)-1]
		}
		m.mu.Unlock()
	}
}

/*

GET www.srv.com/servers/ -> lists all server available

POST www.srv.com/servers/<id>/start
POST www.srv.com/servers/<id>/stop
GET  www.srv.com/servers/<id>/ (websocket connection to logs and commands)



si ya un crash -> email (avec boutton pour restart)

interface web + command line pour acces au terminal + options pour restart stop and start server

add auto build latest version of build tools

opotion to zip and download backup

*/
