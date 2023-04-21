package auth

import (
	"encoding/json"
	"eventapi/config"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Verified  bool   `json:"verified_email"`
	Picture   string `json:"picture"`
	LastName  string `json:"family_name"`
	FirstName string `json:"given_name"`
}

func ConfigGoogle() *oauth2.Config {
	config := config.New()
	conf := &oauth2.Config{
		ClientID:     config.Oauth.Client,
		ClientSecret: config.Oauth.Secret,
		RedirectURL:  config.Oauth.RedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/photoslibrary.readonly.appcreateddata",
			"https://www.googleapis.com/auth/photoslibrary.edit.appcreateddata",
			"https://www.googleapis.com/auth/photoslibrary.appendonly"},
		Endpoint: google.Endpoint,
	}
	return conf
}

//todo: refactor file, possibly move it to a different module

func GetUserInfo(token string) (UserInfo, error) {
	var userInfo UserInfo

	reqURL, err := url.Parse("https://www.googleapis.com/oauth2/v2/userinfo")

	if err != nil {
		return userInfo, err
	}

	ptoken := fmt.Sprintf("Bearer %s", token)
	fmt.Println("token", token)
	res := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Header: map[string][]string{
			"Authorization": {ptoken}},
	}
	req, err := http.DefaultClient.Do(res)
	if err != nil {
		return userInfo, err
	}
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	errorz := json.Unmarshal(body, &userInfo)
	if errorz != nil {
		panic(errorz)
	}
	return userInfo, nil
}
