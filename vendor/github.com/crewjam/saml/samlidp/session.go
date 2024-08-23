package samlidp

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/zenazn/goji/web"

	"github.com/crewjam/saml"
)

var sessionMaxAge = time.Hour

// GetSession returns the *Session for this request.
//
// If the remote user has specified a username and password in the request
// then it is validated against the user database. If valid it sets a
// cookie and returns the newly created session object.
//
// If the remote user has specified invalid credentials then a login form
// is returned with an English-language toast telling the user their
// password was invalid.
//
// If a session cookie already exists and represents a valid session,
// then the session is returned
//
// If neither credentials nor a valid session cookie exist, this function
// sends a login form and returns nil.
func (s *Server) GetSession(w http.ResponseWriter, r *http.Request, req *saml.IdpAuthnRequest) *saml.Session {
	// if we received login credentials then maybe we can create a session
	if r.Method == "POST" && r.PostForm.Get("user") != "" {
		user := User{}
		if err := s.Store.Get(fmt.Sprintf("/users/%s", r.PostForm.Get("user")), &user); err != nil {
			s.sendLoginForm(w, r, req, "Invalid username or password")
			return nil
		}

		if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(r.PostForm.Get("password"))); err != nil {
			s.sendLoginForm(w, r, req, "Invalid username or password")
			return nil
		}

		session := &saml.Session{
			ID:         base64.StdEncoding.EncodeToString(randomBytes(32)),
			NameID:     user.Email,
			CreateTime: saml.TimeNow(),
			ExpireTime: saml.TimeNow().Add(sessionMaxAge),
			Index:      hex.EncodeToString(randomBytes(32)),
			UserName:   user.Name,
			// nolint:gocritic // Groups should be a slice here.
			Groups:                user.Groups[:],
			UserEmail:             user.Email,
			UserCommonName:        user.CommonName,
			UserSurname:           user.Surname,
			UserGivenName:         user.GivenName,
			UserScopedAffiliation: user.ScopedAffiliation,
		}
		if err := s.Store.Put(fmt.Sprintf("/sessions/%s", session.ID), &session); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    session.ID,
			MaxAge:   int(sessionMaxAge.Seconds()),
			HttpOnly: true,
			Secure:   r.URL.Scheme == "https",
			Path:     "/",
		})
		return session
	}

	if sessionCookie, err := r.Cookie("session"); err == nil {
		session := &saml.Session{}
		if err := s.Store.Get(fmt.Sprintf("/sessions/%s", sessionCookie.Value), session); err != nil {
			if err == ErrNotFound {
				s.sendLoginForm(w, r, req, "")
				return nil
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return nil
		}

		if saml.TimeNow().After(session.ExpireTime) {
			s.sendLoginForm(w, r, req, "")
			return nil
		}
		return session
	}

	s.sendLoginForm(w, r, req, "")
	return nil
}

// sendLoginForm produces a form which requests a username and password and directs the user
// back to the IDP authorize URL to restart the SAML login flow, this time establishing a
// session based on the credentials that were provided.
func (s *Server) sendLoginForm(w http.ResponseWriter, _ *http.Request, req *saml.IdpAuthnRequest, toast string) {
	tmpl := template.Must(template.New("saml-post-form").Parse(`` +
		`<html>` +
		`<p>{{.Toast}}</p>` +
		`<form method="post" action="{{.URL}}">` +
		`<input type="text" name="user" placeholder="user" value="" />` +
		`<input type="password" name="password" placeholder="password" value="" />` +
		`<input type="hidden" name="SAMLRequest" value="{{.SAMLRequest}}" />` +
		`<input type="hidden" name="RelayState" value="{{.RelayState}}" />` +
		`<input type="submit" value="Log In" />` +
		`</form>` +
		`</html>`))
	data := struct {
		Toast       string
		URL         string
		SAMLRequest string
		RelayState  string
	}{
		Toast:       toast,
		URL:         req.IDP.SSOURL.String(),
		SAMLRequest: base64.StdEncoding.EncodeToString(req.RequestBuffer),
		RelayState:  req.RelayState,
	}

	if err := tmpl.Execute(w, data); err != nil {
		panic(err)
	}
}

// HandleLogin handles the `POST /login` and `GET /login` forms. If credentials are present
// in the request body, then they are validated. For valid credentials, the response is a
// 200 OK and the JSON session object. For invalid credentials, the HTML login prompt form
// is sent.
func (s *Server) HandleLogin(_ web.C, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	session := s.GetSession(w, r, &saml.IdpAuthnRequest{IDP: &s.IDP})
	if session == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// HandleListSessions handles the `GET /sessions/` request and responds with a JSON formatted list
// of session names.
func (s *Server) HandleListSessions(_ web.C, w http.ResponseWriter, _ *http.Request) {
	sessions, err := s.Store.List("/sessions/")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Sessions []string `json:"sessions"`
	}{Sessions: sessions})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// HandleGetSession handles the `GET /sessions/:id` request and responds with the session
// object in JSON format.
func (s *Server) HandleGetSession(c web.C, w http.ResponseWriter, _ *http.Request) {
	session := saml.Session{}
	err := s.Store.Get(fmt.Sprintf("/sessions/%s", c.URLParams["id"]), &session)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// HandleDeleteSession handles the `DELETE /sessions/:id` request. It invalidates the
// specified session.
func (s *Server) HandleDeleteSession(c web.C, w http.ResponseWriter, _ *http.Request) {
	err := s.Store.Delete(fmt.Sprintf("/sessions/%s", c.URLParams["id"]))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
