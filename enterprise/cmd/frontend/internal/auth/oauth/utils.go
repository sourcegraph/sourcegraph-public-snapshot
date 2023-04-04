package oauth

import (
	"github.com/dghubble/gologin"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func GetStateConfig(name string) gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Name:     name,
		Path:     "/",
		MaxAge:   900, // 15 minutes
		HTTPOnly: true,
		Secure:   conf.IsExternalURLSecure(),
	}
	return cfg
}
