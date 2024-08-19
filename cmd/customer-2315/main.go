package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/google/uuid"
)

type ProxyServer struct {
	accessToken   string
	tokenMutex    sync.RWMutex
	client        *http.Client
	azureEndpoint *url.URL
	logger        log.Logger
}

func (ps *ProxyServer) readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (ps *ProxyServer) generateHeaders(bearerToken string) map[string]string {
	return map[string]string{
		"correlationId":      uuid.New().String(),
		"dataClassification": "sensitive",
		"dataSource":         "internet",
		"Authorization":      "Bearer " + bearerToken,
	}
}

func (ps *ProxyServer) updateAccessToken() {
	for {
		token, err := ps.getAccessToken()
		if err != nil {
			ps.logger.Error("Error getting access token: %v", log.Error(err))
		} else {
			ps.tokenMutex.Lock()
			ps.accessToken = token
			ps.tokenMutex.Unlock()
			ps.logger.Info("Access token updated")
		}
		time.Sleep(1 * time.Minute)
	}
}

func (ps *ProxyServer) initializeAzureEndpoint() {
	var err error
	azure_endpoint, err := ps.readSecretFile("/run/secrets/azure_endpoint")
	if err != nil {
		ps.logger.Fatal("error reading OAUTH_URL: %v", log.Error(err))
	}
	ps.azureEndpoint, err = url.Parse(azure_endpoint)
	if err != nil {
		ps.logger.Fatal("Invalid AZURE_ENDPOINT: %v", log.Error(err))
	}
}

func (ps *ProxyServer) initializeClient() {
	ps.client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        400,
			MaxIdleConnsPerHost: 400,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
		Timeout: 30 * time.Second,
	}
}

func (ps *ProxyServer) getAccessToken() (string, error) {
	url, err := ps.readSecretFile("/run/secrets/oauth_url")
	if err != nil {
		return "", fmt.Errorf("error reading OAUTH_URL: %v", err)
	}
	clientID, err := ps.readSecretFile("/run/secrets/client_id")
	if err != nil {
		return "", fmt.Errorf("error reading CLIENT_ID: %v", err)
	}
	clientSecret, err := ps.readSecretFile("/run/secrets/client_secret")
	if err != nil {
		return "", fmt.Errorf("error reading CLIENT_SECRET: %v", err)
	}

	data := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "azureopenai-readwrite",
		"grant_type":    "client_credentials",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ps.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %v", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return token, nil
}

func (ps *ProxyServer) handleProxy(w http.ResponseWriter, req *http.Request) {
	target := ps.azureEndpoint.ResolveReference(req.URL)
	// Create a proxy request
	proxyReq, err := http.NewRequest(req.Method, target.String(), req.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request
	for header, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	ps.tokenMutex.RLock()
	bearerToken := ps.accessToken
	ps.tokenMutex.RUnlock()
	// Add generated headers
	headers := ps.generateHeaders(bearerToken)
	for key, value := range headers {
		proxyReq.Header.Set(key, value)
	}
	proxyReq.Header.Set("Api-Key", bearerToken)

	resp, err := ps.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Write the headers and status code from the response to the client
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Stream the response body to the client
	reader := bufio.NewReader(resp.Body)
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			ps.logger.Error("Error reading response body: %v", log.Error(err))
			http.Error(w, "Error reading response from upstream server", http.StatusBadGateway)
			return
		}
		if n == 0 {
			break
		}
		if _, writeErr := w.Write(buf[:n]); writeErr != nil {
			ps.logger.Error("Error writing response: %v", log.Error(writeErr))
			break
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func main() {
	liblog := log.Init(log.Resource{
		Name: "Special Oauth Server",
	})
	defer liblog.Sync()

	logger := log.Scoped("server")

	ps := &ProxyServer{
		logger: logger,
	}
	ps.initializeClient()
	ps.initializeAzureEndpoint()
	go ps.updateAccessToken()
	http.HandleFunc("/", ps.handleProxy)
	logger.Info("HTTP Proxy server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("Failed to start HTTP server: %v", log.Error(err))
	}
}
