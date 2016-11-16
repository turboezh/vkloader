package auth

import (
	"fmt"
	"math/rand"
	"net/url"
	"vkloader/util"
)

const REDIRECT_URI = "https://oauth.vk.com/blank.html"
const OAUTH_URL = "https://oauth.vk.com/authorize?display=page&response_type=token&scope=audio,offline"

type Auth struct {
	token  string
	userId string
}

func (a *Auth) OAuthUrl(clientId string) string {
	u, err := url.Parse(OAUTH_URL)
	util.CheckError(err)

	q := u.Query()
	q.Add("client_id", clientId)
	q.Add("redirect_uri", REDIRECT_URI)
	q.Add("state", fmt.Sprint(rand.Uint32()))

	u.RawQuery = q.Encode()

	return u.String()
}

func (a *Auth) ParseAuthURL(authURL string) error {
	u, err := url.Parse(authURL)
	if err != nil {
		return err
	}

	q, err := url.ParseQuery(u.Fragment)
	if err != nil {
		return err
	}

	a.token = q.Get("access_token")
	a.userId = q.Get("user_id")

	return nil
}

func (a *Auth) Token() string {
	return a.token
}

func (a *Auth) SetToken(token string) {
	a.token = token
}

func (a *Auth) UserId() string {
	return a.userId
}

func (a *Auth) SetUserId(userId string) {
	a.userId = userId
}
