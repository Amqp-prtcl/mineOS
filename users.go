package main

import (
	"sync"

	"github.com/Amqp-prtcl/snowflakes"
)

var (
	Users   []*User
	usersmu sync.RWMutex
)

type User struct {
	ID       snowflakes.ID `json:"id"`
	username string        `json:"usr"`
	password string        `json:"pswd"`

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
		if usr.username == usrname {
			return usr, true
		}
	}
	return nil, false
}
