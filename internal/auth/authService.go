package auth

import (
	"context"
	"errors"
	"eventapi/internal/configuration"
	"fmt"

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
	getEntryPoint() string
	exchange(c context.Context, state string, code string) (*oauth2.Token, error)
	verify(c context.Context, token string) (*oidc.IDToken, error)
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
func (s *AuthService) getEntryPoint() string {
	return s.providerConfig.AuthCodeURL(s.cfg.State)
}

func (s *AuthService) exchange(c context.Context, state string, code string) (*oauth2.Token, error) {
	if state != s.cfg.State {
		return nil, errors.New("oauth State did not match request")
	}

	return s.providerConfig.Exchange(c, code)
}
func (s *AuthService) verify(c context.Context, token string) (*oidc.IDToken, error) {
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.cfg.Client})
	return verifier.Verify(c, token)
}
