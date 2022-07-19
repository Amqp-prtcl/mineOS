package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mineOS/config"
	"mineOS/downloads"
	"mineOS/emails"
	"mineOS/manager"
	"mineOS/rooms"
	"mineOS/servers"
	"mineOS/tokens"
	"mineOS/users"
	"mineOS/versions"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 20,
	WriteBufferSize: 1024 * 20,
}

func init() {
	//TODO

	err := config.LoadConfig("/config")
	if err != nil {
		fmt.Printf("[ERR] Unable to load config file.\n")
		panic(err)
	}
	snowflakes.SetEpoch(config.Config.Epoch)

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
	const idRegex = `[0-9]+`
	router := routes.NewRouter(onAuth)

	//SITE
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/login/?$`, NoAuth, getLoginHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/login/?$`, NoAuth, postLoginHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/logout/?$`, NoAuth, logoutHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/$`, Auth, redirectHomeHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/home/?$`, Auth, getHomeHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/?$`, Auth, getServersHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/servers/(`+idRegex+`)/?$`, Auth, getServerHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/servers/(`+idRegex+`)/start/?$`, Auth, startServerHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/servers/(`+idRegex+`)/stop/?$`, Auth, stopServerHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/servers/(`+idRegex+`)/zip/?$`, Auth, zipServerHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/assets/(.+)/?$`, Auth, assetsHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/download/(`+idRegex+`)/?$`, Auth, getDownload))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/download/(`+idRegex+`)/info/?$`, Auth, getDownloadInfo))

	//API
	//general
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/api/epoch/?$`, Auth, getEpochHandler))
	//versions
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/api/versions/?$`, Auth, getsrvTypeListHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/api/versions/([A-Z]+)/?$`, Auth, getVersionIdListHandler))
	//servers
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/api/servers/?$`, Auth, getServerListHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodGet, `^/api/servers/(`+idRegex+`)/?$`, Auth, getServerInfoHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/api/servers/(`+idRegex+`)/emails/?$`, Auth, postServerEmailHandler))
	router.MustAddRoute(routes.MustNewRoute(http.MethodPost, `^/api/servers/new/?$`, Auth, postNewServerHandler))

	//WEBSOCKETS
	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/servers/ws/?$`, Auth, serverListWebsocketHandler))
	router.MustAddRoute(routes.MustNewRoute(routes.HttpMethodAny, `^/servers/(`+idRegex+`)/ws/?$`, Auth, serverWebsocketHandler))

	if err := router.ListenAndServe("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}

/*func TODO(w http.ResponseWriter, r *http.Request, e interface{}, matches []string)*/

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
	http.SetCookie(w, tokens.CookieFromToken(tokens.NewToken(usr)))
	w.WriteHeader(http.StatusNoContent)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.SetCookie(w, tokens.CookieFromToken(nil))
	w.WriteHeader(http.StatusNoContent)
}

func redirectHomeHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.Redirect(w, r, `/servers`, http.StatusPermanentRedirect)
}

func getHomeHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, HomeFile)
}

func getServersHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, RoomsFile)
}

func getServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	http.ServeFile(w, r, RoomFile)
}

func startServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
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
		if err == servers.ErrNotClosed {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func stopServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	err = room.Stop()
	if err != nil {
		if err == servers.ErrNotStarted {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func zipServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
	}
	id, err = room.Zip()
	if err != nil {
		if err == servers.ErrNotClosed {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(struct {
		Id snowflakes.ID `json:"download-id"`
	}{id})
}

func assetsHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	if strings.Contains(matches[0], "..") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join(Assets, matches[0]))
}

func getDownload(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	info, err := downloads.GetInfo(id)
	if err != nil {
		if err == downloads.ErrNoExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dr, err := downloads.GetFile(id)
	if err != nil {
		if err == downloads.ErrNoExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dr.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(info.Name))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	io.Copy(w, dr)
}

func getDownloadInfo(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	info, err := downloads.GetInfo(id)
	if err != nil {
		if err == downloads.ErrNoExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(info)
}

// API

func getEpochHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	json.NewEncoder(w).Encode(struct {
		Epoch time.Time `json:"epoch"`
	}{Epoch: snowflakes.GetEpoch()})
}

func getsrvTypeListHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	json.NewEncoder(w).Encode(versions.GetServerTypes())
}

func getVersionIdListHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	vrs, ok := versions.GetVersionIdsBuServerType(versions.ServerType(matches[0]))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(vrs)
}

func getServerListHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	w.Write(manager.M.MarshalServerList())
}

func getServerInfoHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write(room.MarshalRoomInfo())
}

func postServerEmailHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var remails = []string{}
	err = json.NewDecoder(r.Body).Decode(&remails)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(remails) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var mails = make([]string, len(remails))
	for _, email := range remails {
		mails = append(mails, strings.ToLower(strings.TrimSpace(email)))
	}
	if !emails.AreValidEmails(mails) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room.AddEmail(mails...)
}

func postNewServerHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	var info = struct {
		Name    string              `json:"name"`
		Emails  []string            `json:"emails"`
		SrvType versions.ServerType `json:"server-type"`
		VrsID   string              `json:"version-id"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&info)
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if info.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	prof, err := rooms.GenerateRoom(info.Name, info.SrvType, info.VrsID) // TODO: sanitize upon error (if generation fails on later stage (agreeing to EULA), dead folder will remain on disk -> Must remove it)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ok := manager.M.NewRoom(prof)
	if !ok {
		w.WriteHeader(http.StatusConflict)
		fmt.Printf("??? failed to add new room to roomManager: ID already exist ???")
		return
	}
	json.NewEncoder(w).Encode(struct {
		ID snowflakes.ID `json:"id"`
	}{ID: prof.ID})
}

func serverListWebsocketHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	manager.M.AddConn(conn)
}

func serverWebsocketHandler(w http.ResponseWriter, r *http.Request, e interface{}, matches []string) {
	id, err := snowflakes.ParseID(matches[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	room, ok := manager.M.GetRoombyID(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	room.AddConn(conn)
}
