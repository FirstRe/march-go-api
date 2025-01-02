package constant

import "os"

type OAuthConfig struct {
	ClientID       string
	ClientSecret   string
	AuthURL        string
	TokenURL       string
	RevokeTokenURL string
	RedirectURL    string
}

var ConfigOAuth = OAuthConfig{
	ClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret:   os.Getenv("GOOGLE_SECRET"),
	AuthURL:        "https://accounts.google.com/o/oauth2/v2/auth",
	TokenURL:       "https://oauth2.googleapis.com/token",
	RevokeTokenURL: "https://oauth2.googleapis.com/revoke",
	RedirectURL:    os.Getenv("REDIRECT_URL"),
}
