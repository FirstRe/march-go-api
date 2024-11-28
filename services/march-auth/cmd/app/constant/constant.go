package constant

import "github.com/spf13/viper"

type OAuthConfig struct {
	ClientID       string
	ClientSecret   string
	AuthURL        string
	TokenURL       string
	RevokeTokenURL string
	RedirectURL    string
}

var ConfigOAuth = OAuthConfig{
	ClientID:       viper.GetString("GOOGLE_CLIENT_ID"),
	ClientSecret:   viper.GetString("GOOGLE_SECRET"),
	AuthURL:        "https://accounts.google.com/o/oauth2/v2/auth",
	TokenURL:       "https://oauth2.googleapis.com/token",
	RevokeTokenURL: "https://oauth2.googleapis.com/revoke",
	RedirectURL:    viper.GetString("REDIRECT_URL"),
}
