package webapp

import (
	"net/http"
	"time"
)

func (h *Handler) authCookie(name, value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:    name,
		Value:   value,
		Path:    h.config.ExternalURL.Path,
		Expires: expires,
		// SameSite: http.SameSiteStrictMode, // TODO(sqs): bug in chrome, https://bugs.chromium.org/p/chromium/issues/detail?id=696204
		HttpOnly: true,
	}
}
