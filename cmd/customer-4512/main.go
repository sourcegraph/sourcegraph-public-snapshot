package main

import (
	"bufio"
	"encoding/base64"
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
)

type Proxy struct {
	accessToken   string
	tokenMutex    sync.RWMutex
	client        *http.Client
	azureEndpoint *url.URL
	logger        log.Logger
}

func (ps *Proxy) readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (ps *Proxy) updateAccessToken() {
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

func (ps *Proxy) initializeAzureEndpoint() {
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

func (ps *Proxy) initializeClient() {
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

func (ps *Proxy) getAccessToken() (string, error) {
	oauth_url, err := ps.readSecretFile("/run/secrets/oauth_url")
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

	authKey := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", oauth_url, io.NopCloser(strings.NewReader(data.Encode())))
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Basic "+authKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ps.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to retrieve token: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		ps.logger.Fatal("Failed to unmarshal response body: %v", log.Error(err))
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		ps.logger.Fatal("Failed to retrieve access token from response body")
	}
	return accessToken, nil
}

func (ps *Proxy) handleProxy(w http.ResponseWriter, req *http.Request) {
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

	// Add accesstoken headers
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
		Name: "Cody OAuth Proxy",
	})
	defer liblog.Sync()

	logger := log.Scoped("server")

	ps := &Proxy{logger: logger}
	ps.initializeClient()
	ps.initializeAzureEndpoint()
	go ps.updateAccessToken()
	http.HandleFunc("/", ps.handleProxy)
	logger.Info("HTTP Proxy server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("Failed to start HTTP server: %v", log.Error(err))
	}
}
