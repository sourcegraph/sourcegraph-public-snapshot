package samlidp

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"

	"github.com/zenazn/goji/web"

	"github.com/crewjam/saml"
)

// Service represents a configured SP for whom this IDP provides authentication services.
type Service struct {
	// Name is the name of the service provider
	Name string

	// Metdata is the XML metadata of the service provider.
	Metadata saml.EntityDescriptor
}

// GetServiceProvider returns the Service Provider metadata for the
// service provider ID, which is typically the service provider's
// metadata URL. If an appropriate service provider cannot be found then
// the returned error must be os.ErrNotExist.
func (s *Server) GetServiceProvider(_ *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error) {
	s.idpConfigMu.RLock()
	defer s.idpConfigMu.RUnlock()
	rv, ok := s.serviceProviders[serviceProviderID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return rv, nil
}

// HandleListServices handles the `GET /services/` request and responds with a JSON formatted list
// of service names.
func (s *Server) HandleListServices(_ web.C, w http.ResponseWriter, _ *http.Request) {
	services, err := s.Store.List("/services/")
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		Services []string `json:"services"`
	}{Services: services})
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// HandleGetService handles the `GET /services/:id` request and responds with the service
// metadata in XML format.
func (s *Server) HandleGetService(c web.C, w http.ResponseWriter, _ *http.Request) {
	service := Service{}
	err := s.Store.Get(fmt.Sprintf("/services/%s", c.URLParams["id"]), &service)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = xml.NewEncoder(w).Encode(service.Metadata)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// HandlePutService handles the `PUT /shortcuts/:id` request. It accepts the XML-formatted
// service metadata in the request body and stores it.
func (s *Server) HandlePutService(c web.C, w http.ResponseWriter, r *http.Request) {
	service := Service{}

	metadata, err := getSPMetadata(r.Body)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	service.Metadata = *metadata

	err = s.Store.Put(fmt.Sprintf("/services/%s", c.URLParams["id"]), &service)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s.idpConfigMu.Lock()
	s.serviceProviders[service.Metadata.EntityID] = &service.Metadata
	s.idpConfigMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeleteService handles the `DELETE /services/:id` request.
func (s *Server) HandleDeleteService(c web.C, w http.ResponseWriter, _ *http.Request) {
	service := Service{}
	err := s.Store.Get(fmt.Sprintf("/services/%s", c.URLParams["id"]), &service)
	if err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := s.Store.Delete(fmt.Sprintf("/services/%s", c.URLParams["id"])); err != nil {
		s.logger.Printf("ERROR: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s.idpConfigMu.Lock()
	delete(s.serviceProviders, service.Metadata.EntityID)
	s.idpConfigMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// initializeServices reads all the stored services and initializes the underlying
// identity provider to accept them.
func (s *Server) initializeServices() error {
	serviceNames, err := s.Store.List("/services/")
	if err != nil {
		return err
	}
	for _, serviceName := range serviceNames {
		service := Service{}
		if err := s.Store.Get(fmt.Sprintf("/services/%s", serviceName), &service); err != nil {
			return err
		}

		s.idpConfigMu.Lock()
		s.serviceProviders[service.Metadata.EntityID] = &service.Metadata
		s.idpConfigMu.Unlock()
	}
	return nil
}
