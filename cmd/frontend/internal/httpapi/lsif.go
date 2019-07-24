package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"strings"

	"encoding/hex"
	"encoding/json"

	"fmt"

	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"

	"strconv"
)

func lsifProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

func getLSIFUploadSecret() ([]byte, error) {
	lsifUploadSecret := conf.Get().LsifUploadSecret
	if lsifUploadSecret == "" {
		return nil, errors.New("must set lsifUploadSecret in site config")
	}
	return []byte(lsifUploadSecret), nil
}

func generateUploadToken(repoID api.RepoID, lsifUploadSecret []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(lsifUploadSecret))
	_, err := mac.Write([]byte(strconv.FormatInt(int64(repoID), 10)))
	if err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}

func generateChallenge(userID int32, lsifUploadSecret []byte) string {
	// Must be different from the upload token and different for each user
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", userID, conf.Get().LsifUploadSecret)))
	// The first 10 hex digits is enough to be fairly confident that a GitHub
	// topic of this name doesn't already exist on the repository.
	return hex.EncodeToString(sum[:])[:10]
}

// isValidUploadToken checks whether token is a valid upload token for repoID.
func isValidUploadToken(repoID api.RepoID, tokenString string, lsifUploadSecret []byte) bool {
	tokenBytes, err := hex.DecodeString(tokenString)
	if err != nil {
		return false
	}

	uploadToken, err := generateUploadToken(repoID, lsifUploadSecret)
	if err != nil {
		return false
	}

	return hmac.Equal(tokenBytes, uploadToken)
}

func lsifChallengeHandler(w http.ResponseWriter, r *http.Request) {
	lsifUploadSecret, err := getLSIFUploadSecret()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	actor := actor.FromContext(r.Context())
	json, err := json.Marshal(struct {
		Challenge string `json:"challenge"`
	}{Challenge: generateChallenge(actor.UID, lsifUploadSecret)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(json)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func lsifVerifyHandler(w http.ResponseWriter, r *http.Request) {
	repository := r.URL.Query().Get("repository")
	if !strings.HasPrefix(repository, "github.com") {
		http.Error(w, "Only github.com repositories support verification.", http.StatusUnprocessableEntity)
		return
	}
	ownerAndName := strings.TrimPrefix(repository, "github.com/")

	repo, err := backend.Repos.GetByName(r.Context(), api.RepoName(repository))
	if err != nil {
		http.Error(w, "Unknown repository.", http.StatusUnauthorized)
		return
	}

	actor := actor.FromContext(r.Context())

	apiURL, err := url.Parse("https://api.github.com")
	if err != nil {
		http.Error(w, "Error parsing API URL.", http.StatusInternalServerError)
		return
	}
	client := github.NewClient(apiURL, conf.Get().LsifVerificationGithubToken, nil)
	topics, err := client.ListTopicsOnRepository(r.Context(), ownerAndName)
	if err != nil {
		http.Error(w, "Error listing topics.", http.StatusInternalServerError)
		return
	}

	lsifUploadSecret, err := getLSIFUploadSecret()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	topicsSet := make(map[string]bool)
	for _, v := range topics {
		topicsSet[v] = true
	}
	success := topicsSet[generateChallenge(actor.UID, lsifUploadSecret)]

	var payload interface{}
	if success {
		token, err := generateUploadToken(repo.ID, lsifUploadSecret)
		if err != nil {
			http.Error(w, "Unable to generate LSIF upload token.", http.StatusInternalServerError)
			return
		}
		payload = struct {
			Token string `json:"token"`
		}{Token: hex.EncodeToString(token[:])}
	} else {
		payload = struct {
			Failure string `json:"failure"`
		}{Failure: "Topic not found."}
	}
	json, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(json)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func lsifUploadProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		repo, err := backend.Repos.GetByName(r.Context(), api.RepoName(r.URL.Query().Get("repository")))
		if err != nil {
			http.Error(w, "Unknown repository.", http.StatusUnauthorized)
			return
		}

		lsifUploadSecret, err := getLSIFUploadSecret()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !isValidUploadToken(repo.ID, r.URL.Query().Get("upload_token"), lsifUploadSecret) {
			http.Error(w, "Invalid LSIF upload token.", http.StatusUnauthorized)
			return
		}

		r.URL.Path = "upload"
		p.ServeHTTP(w, r)
	}
}
