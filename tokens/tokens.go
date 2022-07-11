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
	ID    snowflakes.ID
	stamp int64
}

func getTimestamp() int64 {
	return time.Since(snowflakes.GetEpoch()).Milliseconds()
}

func isValidStamp(stamp int64) bool {
	return stamp < getTimestamp()
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
	if !isValidStamp(body.stamp) {
		return nil, nil, false
	}
	usr, ok := users.GetUserbyID(body.ID)
	if !ok {
		return nil, nil, false
	}
	if body.stamp != usr.LastStamp {
		return nil, nil, false
	}
	token = NewToken(usr.ID)
	usr.LastStamp = mustExtractStamp(token)
	return token, usr, true
}

func mustExtractStamp(token jwt.Token) int64 {
	data, err := token.GetBody()
	if err != nil {
		panic(err)
	}
	var body JwtBody
	err = json.Unmarshal(data, &body)
	if err != nil {
		panic(err)
	}
	return body.stamp
}

func NewToken(id snowflakes.ID) jwt.Token {
	data, _ := json.Marshal(&JwtBody{
		ID:    id,
		stamp: getTimestamp() + ExpirationTime.Milliseconds(),
	})
	return jwt.NewToken(data, config.GetSecret())
}

func CookieFromToken(token jwt.Token) *http.Cookie {
	return &http.Cookie{
		Name:     routes.TokenCookieName,
		Value:    token.String(),
		Secure:   false,
		HttpOnly: false,
	}
}