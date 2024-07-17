package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/log"
)

var (
	accessToken   string
	tokenMutex    sync.RWMutex
	client        *http.Client
	azureEndpoint *url.URL
)

func readSecretFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func updateAccessToken(logger log.Logger) {
	for {
		token, err := getAccessToken(logger)
		if err != nil {
			logger.Fatal("Error getting access token: %v", log.Error(err))
		} else {
			tokenMutex.Lock()
			accessToken = token
			tokenMutex.Unlock()
			logger.Info("Access token updated")
		}
		time.Sleep(1 * time.Minute)
	}
}

func initializeAzureEndpoint(logger log.Logger) {
	var err error
	azure_endpoint, err := readSecretFile("/run/secrets/azure_endpoint")
	if err != nil {
		logger.Fatal("error reading OAUTH_URL: %v", log.Error(err))
	}
	azureEndpoint, err = url.Parse(azure_endpoint)
	if err != nil {
		logger.Fatal("Invalid AZURE_ENDPOINT: %v", log.Error(err))
	}
}

func initializeClient() {
	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        400,
			MaxIdleConnsPerHost: 400,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
		Timeout: 30 * time.Second,
	}
}

func getAccessToken(logger log.Logger) (string, error) {
	oauth_url, err := readSecretFile("/run/secrets/oauth_url")
	if err != nil {
		return "", fmt.Errorf("error reading OAUTH_URL: %v", err)
	}
	clientID, err := readSecretFile("/run/secrets/client_id")
	if err != nil {
		return "", fmt.Errorf("error reading CLIENT_ID: %v", err)
	}
	clientSecret, err := readSecretFile("/run/secrets/client_secret")
	if err != nil {
		return "", fmt.Errorf("error reading CLIENT_SECRET: %v", err)
	}

	authKey := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", oauth_url, ioutil.NopCloser(strings.NewReader(data.Encode())))
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Basic "+authKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to retrieve token: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Fatal("Failed to unmarshal response body: %v", log.Error(err))
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		logger.Fatal("Failed to retrieve access token from response body")
	}
	return accessToken, nil
}

func handleProxy(w http.ResponseWriter, req *http.Request, logger log.Logger) {
	target := azureEndpoint.ResolveReference(req.URL)

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

	tokenMutex.RLock()
	bearerToken := accessToken
	tokenMutex.RUnlock()

	// Add accesstoken headers
	proxyReq.Header.Set("Api-Key", bearerToken)
	fmt.Println("the request is made plsease")
	resp, err := client.Do(proxyReq)
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
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
		if n == 0 {
			break
		}
		if _, writeErr := w.Write(buf[:n]); writeErr != nil {
			logger.Fatal("Error writing response: %v", log.Error(writeErr))
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

	initializeClient()
	initializeAzureEndpoint(logger)
	go updateAccessToken(logger)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		handleProxy(w, req, logger)
	})
	logger.Info("HTTPS Proxy server is running on port 8443")
	http.ListenAndServeTLS(":8443", "/run/secrets/cert.pem", "/run/secrets/key.pem", nil)
}
