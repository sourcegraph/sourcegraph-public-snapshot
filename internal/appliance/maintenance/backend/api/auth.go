package api

import (
	"log"
	"net/http"
	"os"
)

const adminKeyCookieName = "admin-password"

var adminPassword = os.Getenv("MAINTENANCE_PASSWORD")

func init() {
	if adminPassword == "" {
		log.Fatal("Variable MAINTENANCE_PASSWORD is missing.")
	}
}

func ensureAuthenticated(endpoint http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !checkAdminKey(r) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			endpoint.ServeHTTP(w, r)
		}
	}
}

func checkAdminKey(r *http.Request) bool {
	if adminKey := r.Header.Get(adminKeyCookieName); adminKey != "" {
		return adminKey == adminPassword
	} else {
		log.Println("Could not find admin key cookie")
		return false
	}
}
