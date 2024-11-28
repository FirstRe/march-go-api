package oAuthService

import (
	"core/app/helper"
	config "march-auth/cmd/app/constant"
	"net/url"

	"github.com/spf13/viper"
)

func OAuthURL() (*string, error) {
	logctx := helper.LogContext("ConfigOAuth", "OAuthURL")

	authParams := url.Values{}
	authParams.Set("client_id", viper.GetString("GOOGLE_CLIENT_ID"))
	authParams.Set("redirect_uri", viper.GetString("REDIRECT_URL"))
	authParams.Set("response_type", "code")
	authParams.Set("scope", "openid profile email")
	authParams.Set("access_type", "offline")
	authParams.Set("state", "standard_oauth")
	authParams.Set("prompt", "consent")

	authURL := config.ConfigOAuth.AuthURL + "?" + authParams.Encode()
	logctx.Logger(config.ConfigOAuth, "ConfigOAuth")
	return &authURL, nil
}
