package main

import (
	"sync"

	"github.com/Amqp-prtcl/snowflakes"
)

var (
	Users []*User = []*User{{
		ID:       snowflakes.NewNode(1).NewID(),
		Username: "Admin",
		Password: "Admin",
	}}
	usersmu sync.RWMutex
)

type User struct {
	ID       snowflakes.ID `json:"id"`
	Username string        `json:"usr"`
	Password string        `json:"pswd"`

	// TODO: add preferences
}

func getUserbyID(id snowflakes.ID) (*User, bool) {
	usersmu.RLock()
	defer usersmu.Unlock()
	for _, usr := range Users {
		if usr.ID == id {
			return usr, true
		}
	}
	return nil, false
}

func getUserbyName(usrname string) (*User, bool) {
	usersmu.RLock()
	defer usersmu.Unlock()
	for _, usr := range Users {
		if usr.Username == usrname {
			return usr, true
		}
	}
	return nil, false
}
