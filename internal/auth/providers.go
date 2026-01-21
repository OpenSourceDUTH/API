package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthConfig holds configuration for all OAuth providers
type OAuthConfig struct {
	Google *oauth2.Config
	GitHub *oauth2.Config
}

// OAuthUserInfo represents user info returned from OAuth providers
type OAuthUserInfo struct {
	ProviderID  string
	Email       string
	DisplayName string
}

// ProviderConfig holds the credentials for an OAuth provider
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
}

// NewOAuthConfig creates OAuth configurations for all providers
func NewOAuthConfig(googleCfg, githubCfg ProviderConfig, callbackBaseURL string) *OAuthConfig {
	config := &OAuthConfig{}

	if googleCfg.ClientID != "" && googleCfg.ClientSecret != "" {
		config.Google = &oauth2.Config{
			ClientID:     googleCfg.ClientID,
			ClientSecret: googleCfg.ClientSecret,
			RedirectURL:  callbackBaseURL + "/api/auth/callback/google",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	if githubCfg.ClientID != "" && githubCfg.ClientSecret != "" {
		config.GitHub = &oauth2.Config{
			ClientID:     githubCfg.ClientID,
			ClientSecret: githubCfg.ClientSecret,
			RedirectURL:  callbackBaseURL + "/api/auth/callback/github",
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: github.Endpoint,
		}
	}

	return config
}

// GetAuthURL returns the OAuth authorization URL for a provider
func (c *OAuthConfig) GetAuthURL(provider Provider, state string) (string, error) {
	cfg, err := c.getConfig(provider)
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// ExchangeCode exchanges an authorization code for tokens
func (c *OAuthConfig) ExchangeCode(ctx context.Context, provider Provider, code string) (*oauth2.Token, error) {
	cfg, err := c.getConfig(provider)
	if err != nil {
		return nil, err
	}
	return cfg.Exchange(ctx, code)
}

// GetUserInfo fetches user information from the OAuth provider
func (c *OAuthConfig) GetUserInfo(ctx context.Context, provider Provider, token *oauth2.Token) (*OAuthUserInfo, error) {
	cfg, err := c.getConfig(provider)
	if err != nil {
		return nil, err
	}

	client := cfg.Client(ctx, token)

	switch provider {
	case ProviderGoogle:
		return c.getGoogleUserInfo(client)
	case ProviderGitHub:
		return c.getGitHubUserInfo(client)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (c *OAuthConfig) getConfig(provider Provider) (*oauth2.Config, error) {
	switch provider {
	case ProviderGoogle:
		if c.Google == nil {
			return nil, fmt.Errorf("google OAuth not configured")
		}
		return c.Google, nil
	case ProviderGitHub:
		if c.GitHub == nil {
			return nil, fmt.Errorf("github OAuth not configured")
		}
		return c.GitHub, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// GoogleUserInfo represents Google's userinfo response
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (c *OAuthConfig) getGoogleUserInfo(client *http.Client) (*OAuthUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API error: %s", string(body))
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Email == "" {
		return nil, fmt.Errorf("email not provided by Google")
	}

	displayName := info.Name
	if displayName == "" {
		displayName = info.Email
	}

	return &OAuthUserInfo{
		ProviderID:  info.ID,
		Email:       info.Email,
		DisplayName: displayName,
	}, nil
}

// GitHubUserInfo represents GitHub's user response
type GitHubUserInfo struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubEmail represents a GitHub email response
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (c *OAuthConfig) getGitHubUserInfo(client *http.Client) (*OAuthUserInfo, error) {
	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API error: %s", string(body))
	}

	var info GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// If email is empty, fetch from emails endpoint
	email := info.Email
	if email == "" {
		email, err = c.getGitHubPrimaryEmail(client)
		if err != nil {
			return nil, err
		}
	}

	displayName := info.Name
	if displayName == "" {
		displayName = info.Login
	}

	return &OAuthUserInfo{
		ProviderID:  fmt.Sprintf("%d", info.ID),
		Email:       email,
		DisplayName: displayName,
	}, nil
}

func (c *OAuthConfig) getGitHubPrimaryEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github emails API error: %s", string(body))
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find primary verified email
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	// Fallback to any verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}

// IsProviderConfigured checks if a provider is configured
func (c *OAuthConfig) IsProviderConfigured(provider Provider) bool {
	switch provider {
	case ProviderGoogle:
		return c.Google != nil
	case ProviderGitHub:
		return c.GitHub != nil
	default:
		return false
	}
}
