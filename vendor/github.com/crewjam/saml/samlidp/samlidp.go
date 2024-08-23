// Package samlidp a rudimentary SAML identity provider suitable for
// testing or as a starting point for a more complex service.
package samlidp

import (
	"crypto"
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/zenazn/goji/web"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
)

// Options represent the parameters to New() for creating a new IDP server
type Options struct {
	URL         url.URL
	Key         crypto.PrivateKey
	Signer      crypto.Signer
	Logger      logger.Interface
	Certificate *x509.Certificate
	Store       Store
}

// Server represents an IDP server. The server provides the following URLs:
//
//	/metadata     - the SAML metadata
//	/sso          - the SAML endpoint to initiate an authentication flow
//	/login        - prompt for a username and password if no session established
//	/login/:shortcut - kick off an IDP-initiated authentication flow
//	/services     - RESTful interface to Service objects
//	/users        - RESTful interface to User objects
//	/sessions     - RESTful interface to Session objects
//	/shortcuts    - RESTful interface to Shortcut objects
type Server struct {
	http.Handler
	idpConfigMu      sync.RWMutex // protects calls into the IDP
	logger           logger.Interface
	serviceProviders map[string]*saml.EntityDescriptor
	IDP              saml.IdentityProvider // the underlying IDP
	Store            Store                 // the data store
}

// New returns a new Server
func New(opts Options) (*Server, error) {
	metadataURL := opts.URL
	metadataURL.Path += "/metadata"
	ssoURL := opts.URL
	ssoURL.Path += "/sso"
	logr := opts.Logger
	if logr == nil {
		logr = logger.DefaultLogger
	}

	s := &Server{
		serviceProviders: map[string]*saml.EntityDescriptor{},
		IDP: saml.IdentityProvider{
			Key:         opts.Key,
			Signer:      opts.Signer,
			Logger:      logr,
			Certificate: opts.Certificate,
			MetadataURL: metadataURL,
			SSOURL:      ssoURL,
		},
		logger: logr,
		Store:  opts.Store,
	}

	s.IDP.SessionProvider = s
	s.IDP.ServiceProviderProvider = s

	if err := s.initializeServices(); err != nil {
		return nil, err
	}
	s.InitializeHTTP()
	return s, nil
}

// InitializeHTTP sets up the HTTP handler for the server. (This function
// is called automatically for you by New, but you may need to call it
// yourself if you don't create the object using New.)
func (s *Server) InitializeHTTP() {
	mux := web.New()
	s.Handler = mux

	mux.Get("/metadata", func(w http.ResponseWriter, r *http.Request) {
		s.idpConfigMu.RLock()
		defer s.idpConfigMu.RUnlock()
		s.IDP.ServeMetadata(w, r)
	})
	mux.Handle("/sso", func(w http.ResponseWriter, r *http.Request) {
		s.idpConfigMu.RLock()
		defer s.idpConfigMu.RUnlock()
		s.IDP.ServeSSO(w, r)
	})

	mux.Handle("/login", s.HandleLogin)
	mux.Handle("/login/:shortcut", s.HandleIDPInitiated)
	mux.Handle("/login/:shortcut/*", s.HandleIDPInitiated)

	mux.Get("/services/", s.HandleListServices)
	mux.Get("/services/:id", s.HandleGetService)
	mux.Put("/services/:id", s.HandlePutService)
	mux.Post("/services/:id", s.HandlePutService)
	mux.Delete("/services/:id", s.HandleDeleteService)

	mux.Get("/users/", s.HandleListUsers)
	mux.Get("/users/:id", s.HandleGetUser)
	mux.Put("/users/:id", s.HandlePutUser)
	mux.Delete("/users/:id", s.HandleDeleteUser)

	sessionPath := regexp.MustCompile("/sessions/(?P<id>.*)")
	mux.Get("/sessions/", s.HandleListSessions)
	mux.Get(sessionPath, s.HandleGetSession)
	mux.Delete(sessionPath, s.HandleDeleteSession)

	mux.Get("/shortcuts/", s.HandleListShortcuts)
	mux.Get("/shortcuts/:id", s.HandleGetShortcut)
	mux.Put("/shortcuts/:id", s.HandlePutShortcut)
	mux.Delete("/shortcuts/:id", s.HandleDeleteShortcut)
}
