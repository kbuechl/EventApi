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

type AuthService struct {
	cfg            *configuration.OauthConfig
	providerConfig *oauth2.Config
}

type oAuthService interface {
	getEntryPoint() string
	exchange(c context.Context, state string, code string) (*oauth2.Token, error)
}

func NewAuthService(cfg *configuration.OauthConfig) (*AuthService, error) {
	if cfg.Client == "" || cfg.Secret == "" || cfg.State == "" {
		return nil, fmt.Errorf("failed to initialize auth provider: missing critical config values")
	}

	return &AuthService{
		cfg: cfg,
		providerConfig: &oauth2.Config{
			ClientID:     cfg.Client,
			ClientSecret: cfg.Secret,
			RedirectURL:  cfg.RedirectUrl,
			Scopes: []string{
				//todo can we move to incremental scope requests?
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/photoslibrary.readonly.appcreateddata",
				"https://www.googleapis.com/auth/photoslibrary.edit.appcreateddata",
				"https://www.googleapis.com/auth/photoslibrary.appendonly"},
			Endpoint: google.Endpoint,
		},
	}, nil
}

// get entry point for starting oauth flow
func (s *AuthService) getEntryPoint() string {
	return s.providerConfig.AuthCodeURL(s.cfg.State)
}

func (s *AuthService) exchange(c context.Context, state string, code string) (*oauth2.Token, error) {
	if state != s.cfg.State {
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
