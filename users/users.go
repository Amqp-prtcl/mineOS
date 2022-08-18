package users

import (
	"encoding/json"
	"fmt"
	"mineOS/globals"
	"os"
	"sync"

	"github.com/Amqp-prtcl/snowflakes"
)

var (
	Users     []*User = []*User{}
	usersmu   sync.RWMutex
	UsersNode = snowflakes.NewNode(1)
)

type User struct {
	ID       snowflakes.ID `json:"id"`
	Username string        `json:"usr"`
	Password string        `json:"pswd"`

	LastStamp int64 `json:"-"`
}

func newDefaultUser() *User {
	return &User{
		ID:       UsersNode.NewID(),
		Username: "Admin",
		Password: "Admin",
	}
}

// if file arg is not present, it is fetched from config file
func LoadUsers(file string) error {
	if file == "" {
		file = globals.UsersFile.WarnGet()
	}
	usersmu.Lock()
	defer usersmu.Unlock()
	if len(Users) != 0 {
		return fmt.Errorf("double loading of users")
	}
	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			Users = []*User{}
			fmt.Printf("Users File not found, created new default admin user: {username: Admin, password: Admin}\n")
			Users = append(Users, newDefaultUser())
			return nil
		}
		return err
	}
	Users = []*User{}
	err = json.NewDecoder(f).Decode(&Users)
	f.Close()
	if err != nil {
		return err
	}

	if len(Users) == 0 {
		fmt.Printf("No user found, created new default admin user: {username: Admin, password: Admin}\n")
		Users = append(Users, newDefaultUser())
	}
	return nil
}

func GetUserbyID(id snowflakes.ID) (*User, bool) {
	usersmu.RLock()
	defer usersmu.RUnlock()
	for _, usr := range Users {
		if usr.ID == id {
			return usr, true
		}
	}
	return nil, false
}

func GetUserbyName(usrname string) (*User, bool) {
	usersmu.RLock()
	defer usersmu.RUnlock()
	for _, usr := range Users {
		if usr.Username == usrname {
			return usr, true
		}
	}
	return nil, false
}
