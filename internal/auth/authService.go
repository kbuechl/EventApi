package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"eventapi/internal/configuration"
	"eventapi/internal/database"
	"fmt"
	"io"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oidcUrl = "https://accounts.google.com"

type AuthService struct {
	cfg            *configuration.OauthConfig
	providerConfig *oauth2.Config
	provider       *oidc.Provider
}

type oAuthService interface {
	getEntryPoint(nonce string) string
	exchange(c context.Context, state string, code string) (*oauth2.Token, error)
	verify(c context.Context, token string) (*oidc.IDToken, error)
	getUserInfo(c context.Context, token *oauth2.Token) (*database.OidcUser, error)
}

func NewAuthService(cfg *configuration.OauthConfig) (*AuthService, error) {
	if cfg.Client == "" || cfg.Secret == "" || cfg.State == "" {
		return nil, fmt.Errorf("failed to initialize auth provider: missing critical config values")
	}

	provider, err := oidc.NewProvider(context.Background(), oidcUrl)

	if err != nil {
		return nil, fmt.Errorf("error creating oidc provider: %w", err)
	}

	return &AuthService{
		cfg:      cfg,
		provider: provider,
		providerConfig: &oauth2.Config{
			ClientID:     cfg.Client,
			ClientSecret: cfg.Secret,
			RedirectURL:  cfg.RedirectUrl,
			Scopes: []string{
				oidc.ScopeOpenID,
				"email", "profile",
				//todo can we move to incremental scope requests?
				"https://www.googleapis.com/auth/photoslibrary.readonly.appcreateddata",
				"https://www.googleapis.com/auth/photoslibrary.edit.appcreateddata",
				"https://www.googleapis.com/auth/photoslibrary.appendonly"},
			Endpoint: google.Endpoint,
		},
	}, nil
}

// get entry point for starting oauth flow
func (s *AuthService) getEntryPoint(nonce string) string {
	return s.providerConfig.AuthCodeURL(s.cfg.State, oidc.Nonce(nonce))
}

func (s *AuthService) exchange(c context.Context, state string, code string) (*oauth2.Token, error) {
	if state != s.cfg.State {
		return nil, errors.New("oauth State did not match request")
	}

	return s.providerConfig.Exchange(c, code, oauth2.AccessTypeOffline)
}
func (s *AuthService) verify(c context.Context, token string) (*oidc.IDToken, error) {
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.cfg.Client})
	return verifier.Verify(c, token)
}
func (s *AuthService) getUserInfo(c context.Context, token *oauth2.Token) (*database.OidcUser, error) {
	u, err := s.provider.UserInfo(c, oauth2.StaticTokenSource(token))

	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}
	var userInfo = &database.OidcUser{}

	if err := u.Claims(userInfo); err != nil {
		return nil, fmt.Errorf("error getting user claim: %w", err)
	}

	return userInfo, nil
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
