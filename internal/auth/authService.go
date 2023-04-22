package auth

import (
	"context"
	"encoding/json"
	"eventapi/config"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var configuration *config.Config
var providerConfig *oauth2.Config

func init() {
	configuration = config.New()
	providerConfig = &oauth2.Config{
		ClientID:     configuration.Oauth.Client,
		ClientSecret: configuration.Oauth.Secret,
		RedirectURL:  configuration.Oauth.RedirectUrl,
		Scopes: []string{
			//todo can we move to incremental scope requests?
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/photoslibrary.readonly.appcreateddata",
			"https://www.googleapis.com/auth/photoslibrary.edit.appcreateddata",
			"https://www.googleapis.com/auth/photoslibrary.appendonly"},
		Endpoint: google.Endpoint,
	}
}

// get entry point for starting oauth flow
func getEntryPoint() string {
	return providerConfig.AuthCodeURL(configuration.Oauth.State)
}

func exchange(c context.Context, state string, code string) (*oauth2.Token, error) {
	if state != configuration.Oauth.State {
		panic("Oauth State did not match request")
	}

	return providerConfig.Exchange(c, code)
}

// Get user info from oauth provider with user token
func getUserInfo(t string) (*UserInfo, error) {
	var userInfo UserInfo

	reqURL, err := url.Parse(userInfoURL)

	if err != nil {
		return nil, err
	}

	ptoken := fmt.Sprintf("Bearer %s", t)
	res := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Header: map[string][]string{
			"Authorization": {ptoken}},
	}
	req, err := http.DefaultClient.Do(res)

	if err != nil {
		return nil, err
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
