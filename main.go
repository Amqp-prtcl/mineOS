package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mineOS/config"
	"mineOS/manager"
	"mineOS/rooms"
	"mineOS/tokens"
	"mineOS/users"
	"mineOS/versions"
	"net/http"
	"os"
	"os/exec"
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
	Root      = "/mineos/data/"
	Assets    = Root + "assets/"
	UsersFile = Root + "users.json"

	LoginFile   = Assets + "login.html"
	HomeFile    = Assets + "home.html"
	RoomsFile   = Assets + "rooms.html"
	RoomFile    = Assets + "room.html"
	NewRoomFile = Assets + "newRoom.html"

	Epoch time.Time = time.UnixMicro(0) //TODO
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 20,
	WriteBufferSize: 1024 * 20,
}

func init() {
	snowflakes.SetEpoch(Epoch)
	//TODO

	err := config.LoadConfig("/config")
	if err != nil {
		fmt.Printf("[ERR] Unable to load config file.\n")
		panic(err)
	}

	//protocol:
	// create directories
	/*err := os.MkdirAll(Assets, 0664)
	if err != nil {
		panic(err)
	}*/
	info, err := os.Stat(Assets)
	if err != nil || info.IsDir() {
		fmt.Printf("[ERR] Assest directory not found\n")
		panic(err)
	}

	// check for users -> if no users create admin and ask to change default password
	//TODO: load users from file
	err = users.LoadUsers("")
	if err != nil {
		panic(err)
	}

	//load servers -> load manager
	err = manager.M.LoadRooms("")
	if err != nil {
		fmt.Printf("[ERR] failed to load servers profile file.\n")
		panic(err)
	}

	// check for java
	fmt.Printf("checking for java... ")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	err = exec.CommandContext(ctx, "java", "-version").Run()
	cancel()
	if err != nil {
		fmt.Printf("\n[ERR] Java not Found\n")
		panic(err)
	}
	fmt.Printf("java found\n")
	//TODO: check git version

	//fetching minecraft vanilla versions
	fmt.Printf("fetching minecraft versions...\n")
	err = versions.Setup()
	if err != nil {
		fmt.Printf("failed to fetch vanilla versions...\n")
		panic(err)
	}

	fmt.Printf("starting app...\n")
}

func onAuth(r *http.Request, authType int, token jwt.Token) (*http.Cookie, interface{}, bool) {
	if authType == NoAuth {
		return nil, nil, true
	}
	token, usr, ok := tokens.ProcessToken(token)
	if !ok {
		return nil, nil, false
	}
	return tokens.CookieFromToken(token), usr, true
}

func main() {
	router := routes.NewRouter(onAuth)

	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/?$`, Auth, homeHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/login/?(.*)/?$`, NoAuth, getLoginHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/assets/(.+)/?$`, Auth, assetsHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/login/?(.*)/?$`, NoAuth, postLoginHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/?$`, Auth, getRoomsHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/ls/?$`, Auth, listServerHandler))
	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/servers/ws/?$`, Auth, RoomsSocketHandler))

	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/servers/new/?$`, Auth, postNewServerHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/new/?$`, Auth, getNewServerHandler))

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
	usr, ok := users.GetUserbyName(r.PostFormValue("username"))
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if usr.Password != r.PostFormValue("password") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, tokens.CookieFromToken(tokens.NewToken(usr.ID)))
	w.WriteHeader(http.StatusNoContent)
}

func getRoomsHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, RoomsFile)
}

func listServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	w.Write(manager.M.MarshalServerList())
}

func RoomsSocketHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	manager.M.AddConn(conn)
}

func postNewServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	var body = &struct {
		Name       string              `json:"name"`
		ServerType versions.ServerType `json:"server-type"`
		VersionID  string              `json:"version-id"`
		Emails     []string            `jdon:"emails"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.ServerType == "" || body.VersionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//TODO

	profile, err := rooms.GenerateRoom(body.Name, body.ServerType, body.VersionID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	profile.Emails = body.Emails
	ok := manager.M.NewRoom(profile)
	if !ok { // should never trigger since a new token is guaranteed to be unique
		fmt.Printf("??? failed to add new room to roomManager: ID already existing ???")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getNewServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, NewRoomFile)
}

func getRoomHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, RoomFile)
}

func RoomSocketHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	room.AddConn(conn)
}

func startRoomHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	err = room.Start()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func stopRoomHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	room.Stop()
	w.WriteHeader(http.StatusNoContent)
}

func assetsHandler(w http.ResponseWriter, r *http.Request, entity interface{}, matches []string) {
	if strings.Contains(matches[0], "..") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(Root, "assets", matches[0]))
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
