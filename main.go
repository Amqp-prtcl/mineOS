package main

import (
	"fmt"
	"log"
	"mineOS/servers"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Amqp-prtcl/jwt"
	"github.com/Amqp-prtcl/routes"
	"github.com/Amqp-prtcl/snowflakes"
	"github.com/gorilla/websocket"
)

const (
	NoAuth = iota
	Auth
)

var (
	server *servers.Server
	Root   = "/mineos/data/"
	Assets = Root + "assets/"

	LoginFile = Assets + "login.html"
	HomeFile  = Assets + "home.html"
	RoomsFile = Assets + "rooms.html"
	RoomFile  = Assets + "room.html"

	Secret = "//TODO"
	Epoch  time.Time //TODO
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 20,
	WriteBufferSize: 1024 * 20,
}

func onAuth(r *http.Request, authType int, token jwt.Token) (*http.Cookie, interface{}, bool) {
	if authType == NoAuth {
		return nil, nil, true
	}
	token, usr, ok := processToken(token)
	if !ok {
		return nil, nil, false
	}
	return CookieFromToken(token), usr, true
}

func main() {
	snowflakes.SetEpoch(Epoch)
	router := routes.NewRouter(onAuth)

	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/?$`, Auth, homeHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/login/?(.*)/?$`, NoAuth, getLoginHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/assets/(.+)/?$`, Auth, assetsHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/login/?(.*)/?$`, NoAuth, postLoginHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/?$`, Auth, getRoomsHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/ls/?$`, Auth, listServerHandler))
	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/servers/ws/?$`, Auth, RoomsSocketHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/(.+)/?$`, Auth, getRoomHandler))
	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/servers/(.+)/ws/?$`, Auth, RoomSocketHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/servers/(.+)/start/?$`, Auth, startRoomHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/servers/(.+)/stop/?$`, Auth, stopRoomHandler))

	if err := router.ListenAndServe("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}

/*

GET  http://server.com/servers/
GET  http://server.com/servers/<id>/
POST http://server.com/servers/<id>/start/
POST http://server.com/servers/<id>/stop/
any  ws://server.com/servers/<id>/ws

Serve files:
GET http://server.com/assets/...

*/

func homeHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	//http.ServeFile(w, r, HomeFile)
	http.Redirect(w, r, "/servers", http.StatusPermanentRedirect)
}

func getLoginHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, LoginFile)
}

func postLoginHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	usr, ok := getUserbyName(r.PostFormValue("username"))
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if usr.password != r.PostFormValue("password") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, CookieFromToken(NewToken(usr.ID)))
	w.WriteHeader(http.StatusNoContent)
}

func getRoomsHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, RoomsFile)
}

func listServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	//TODO
}

func RoomsSocketHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	//TODO
	/*
		if strings.HasSuffix(path, "/ws") {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		manager.addConn(conn)
		return
	*/
}

func getRoomHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, RoomFile)
}

func RoomSocketHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	//TODO
}

func startRoomHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	//TODO
}

func stopRoomHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	//TODO
}

func assetsHandler(w http.ResponseWriter, r *http.Request, entity interface{}, matches []string) {
	if strings.Contains(matches[0], "..") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(Root, "assets", matches[0]))
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
