package tokens

import (
	"encoding/json"
	"mineOS/config"
	"mineOS/users"
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
	ID    snowflakes.ID `json:"id"`
	Stamp int64         `json:"stamp"`
}

func getTimestamp() int64 {
	return time.Since(snowflakes.GetEpoch()).Milliseconds()
}

func isValidStamp(stamp int64) bool {
	return stamp > getTimestamp()
}

func ProcessToken(token jwt.Token) (jwt.Token, *users.User, bool) {
	if token == nil {
		return nil, nil, false
	}
	data, ok := token.ValidateToken(config.GetSecret())
	if !ok {
		return nil, nil, false
	}
	var body = new(JwtBody)
	err := json.Unmarshal(data, body)
	if err != nil {
		return nil, nil, false
	}
	if !isValidStamp(body.Stamp) {
		return nil, nil, false
	}
	usr, ok := users.GetUserbyID(body.ID)
	if !ok {
		return nil, nil, false
	}
	if body.Stamp != usr.LastStamp {
		return nil, nil, false
	}
	token = NewToken(usr)
	return token, usr, true
}

func NewToken(usr *users.User) jwt.Token {
	usr.LastStamp = getTimestamp() + ExpirationTime.Milliseconds()
	data, _ := json.Marshal(&JwtBody{
		ID:    usr.ID,
		Stamp: usr.LastStamp,
	})
	return jwt.NewToken(data, config.GetSecret())
}

func CookieFromToken(token jwt.Token) *http.Cookie {
	var val string
	if token != nil {
		val = token.String()
	}
	return &http.Cookie{
		Name:     routes.TokenCookieName,
		Value:    val,
		Secure:   false,
		HttpOnly: false,
		Path:     "/",
	}
}
