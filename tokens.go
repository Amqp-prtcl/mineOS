package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Amqp-prtcl/jwt"
	"github.com/Amqp-prtcl/routes"
	"github.com/Amqp-prtcl/snowflakes"
)

var (
	ExpirationTime = time.Hour
)

type JwtBody struct {
	ID    snowflakes.ID
	stamp int64
}

func getTimestamp() int64 {
	return time.Since(snowflakes.GetEpoch()).Milliseconds()
}

func isValidStamp(stamp int64) bool {
	return stamp < getTimestamp()
}

func processToken(token jwt.Token) (jwt.Token, *User, bool) {
	if token == nil {
		return nil, nil, false
	}
	data, ok := token.ValidateToken(Secret)
	if !ok {
		return nil, nil, false
	}
	var body = new(JwtBody)
	err := json.Unmarshal(data, body)
	if err != nil {
		return nil, nil, false
	}
	if !isValidStamp(body.stamp) {
		return nil, nil, false
	}
	usr, ok := getUserbyID(body.ID)
	if !ok {
		return nil, nil, false
	}
	return NewToken(usr.ID), usr, true
}

func NewToken(id snowflakes.ID) jwt.Token {
	data, _ := json.Marshal(&JwtBody{
		ID:    id,
		stamp: getTimestamp() + ExpirationTime.Milliseconds(),
	})
	return jwt.NewToken(data, Secret)
}

func CookieFromToken(token jwt.Token) *http.Cookie {
	return &http.Cookie{
		Name:     routes.TokenCookieName,
		Value:    token.String(),
		Secure:   false,
		HttpOnly: false,
	}
}
