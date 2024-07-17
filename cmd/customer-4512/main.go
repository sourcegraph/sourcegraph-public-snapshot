package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

func generateHeaders(bearerToken string) map[string]string {
	return map[string]string{
		"correlationId":      uuid.New().String(),
		"dataClassification": "sensitive",
		"dataSource":         "internet",
		"Authorization":      "Bearer " + bearerToken,
	}
}

func updateAccessToken() {
	for {
		token, err := getAccessToken()
		if err != nil {
			log.Printf("Error getting access token: %v", err)
		} else {
			tokenMutex.Lock()
			accessToken = token
			tokenMutex.Unlock()
			log.Println("Access token updated")
		}
		time.Sleep(1 * time.Minute)
	}
}

func initializeAzureEndpoint() {
	var err error
	azure_endpoint, err := readSecretFile("/run/secrets/azure_endpoint")
	if err != nil {
		log.Fatalf("error reading OAUTH_URL: %v", err)
	}
	azureEndpoint, err = url.Parse(azure_endpoint)
	if err != nil {
		log.Fatalf("Invalid AZURE_ENDPOINT: %v", err)
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

func getAccessToken() (string, error) {
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
		log.Fatalf("Failed to unmarshal response body: %v", err)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		log.Fatalf("Failed to retrieve access token from response body")
	}
	return accessToken, nil
}

func handleProxy(w http.ResponseWriter, req *http.Request) {
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
			log.Printf("Error writing response: %v", writeErr)
			break
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func main() {
	initializeClient()
	initializeAzureEndpoint()
	go updateAccessToken()
	http.HandleFunc("/", handleProxy)
	log.Println("HTTPS Proxy server is running on port 8443")
	log.Fatal(http.ListenAndServeTLS(":8443", "/run/secrets/cert.pem", "/run/secrets/key.pem", nil))
}
