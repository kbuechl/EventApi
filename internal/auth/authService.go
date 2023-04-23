package auth

import (
	"context"
	"encoding/json"
	"errors"
	"eventapi/internal/configuration"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type authConfig struct {
	Secret      string
	Client      string
	RedirectUrl string
	State       string
}

type AuthService struct {
	config         *authConfig
	providerConfig *oauth2.Config
}

type oAuthService interface {
	getEntryPoint() string
	exchange(c context.Context, state string, code string) (*oauth2.Token, error)
}

func NewAuthService() *AuthService {
	c := configure()

	return &AuthService{
		config: c,
		providerConfig: &oauth2.Config{
			ClientID:     c.Client,
			ClientSecret: c.Secret,
			RedirectURL:  c.RedirectUrl,
			Scopes: []string{
				//todo can we move to incremental scope requests?
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/photoslibrary.readonly.appcreateddata",
				"https://www.googleapis.com/auth/photoslibrary.edit.appcreateddata",
				"https://www.googleapis.com/auth/photoslibrary.appendonly"},
			Endpoint: google.Endpoint,
		},
	}

}

// get entry point for starting oauth flow
func (s *AuthService) getEntryPoint() string {
	return s.providerConfig.AuthCodeURL(s.config.State)
}

func (s *AuthService) exchange(c context.Context, state string, code string) (*oauth2.Token, error) {
	if state != s.config.State {
		return nil, errors.New("oauth State did not match request")
	}

	return s.providerConfig.Exchange(c, code)
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

func configure() *authConfig {
	secret, err := configuration.GetRequiredEnv("OAUTH_SECRET")
	if err != nil {
		panic(err)
	}
	c, err := configuration.GetRequiredEnv("OAUTH_CLIENT")
	if err != nil {
		panic(err)
	}
	state, err := configuration.GetRequiredEnv("OAUTH_STATE")
	if err != nil {
		panic(err)
	}

	return &authConfig{
		Secret:      secret,
		Client:      c,
		RedirectUrl: configuration.GetEnv("REDIRECT_URL", "/"),
		State:       state,
	}
}
