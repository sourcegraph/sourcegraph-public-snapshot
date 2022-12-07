package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var (
	copilotProxyAddress = flag.String("remote", "", "http(s)://${host}:${port} of the copilot proxy server. The official one is https://copilot-proxy.githubusercontent.com.")
	myAddress           = flag.String("addr", "127.0.0.1:5000", "The addr of the application.")
)

func main() {
	flag.Parse()

	if *copilotProxyAddress == "" {
		log.Fatal("Must specific --remote")
	}

	handler := &proxy{}
	log.Printf("Starting proxy server: %s -> %s", *myAddress, *copilotProxyAddress)
	if err := http.ListenAndServe(*myAddress, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

type proxy struct{}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	if err := p.serveHTTP(wr, req); err != nil {
		log.Printf("Unhandled error: %s", err)
		http.Error(wr, err.Error(), http.StatusInternalServerError)
	}
}

func (p *proxy) serveHTTP(wr http.ResponseWriter, req *http.Request) error {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	var jBody map[string]interface{}
	if err := json.Unmarshal(body, &jBody); err != nil {
		return err
	}

	// Truncate to 1024 characters for now, necessary for some reason,
	// otherwise you get garbage
	const maxLen = 1024
	if prompt, ok := jBody["prompt"].(string); ok && len(prompt) > maxLen {
		log.Printf("Truncating to maxLen: %s", prompt)
		jBody["prompt"] = prompt[len(prompt)-maxLen:]
	}

	bodyPretty, err := json.MarshalIndent(jBody, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("REQ: %s %s\n%s", req.Method, req.URL, string(bodyPretty))
	bodyNew, err := json.Marshal(jBody)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req.RequestURI = ""

	newURLStr := fmt.Sprintf("%s%s", *copilotProxyAddress, req.URL.Path)
	if queryStr := req.URL.Query().Encode(); queryStr != "" {
		newURLStr += queryStr
	}
	if frag := req.URL.Fragment; frag != "" {
		newURLStr += frag
	}
	newReq, err := http.NewRequest(req.Method, newURLStr, bytes.NewReader(bodyNew))
	if err != nil {
		return errors.Wrap(err, "http.NewRequest")
	}
	newReq.Header = req.Header

	resp, err := client.Do(newReq)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		return errors.Wrap(err, "client.Do")
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadAll")
	}

	respBodyStr, didCleanUp := cleanupResponse(string(respBody))
	if didCleanUp {
		log.Print("[WARN] Modified response to clean up extraneous \"data:\" substrings")
	}

	log.Printf("########### RESP: %s\n%s\n###########", resp.Status, respBodyStr)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, strings.NewReader(respBodyStr))
	return nil
}

// Clean up the response on streaming requests sent by Copilot client
// (this is likely due to a bug in the Fauxpilot Copilot proxy)
func cleanupResponse(body string) (cleanedUp string, didCleanUp bool) {
	if !strings.Contains(body, "data:") {
		return body, false
	}
	body = strings.TrimPrefix(body, "data: data: ")
	for i := 0; i < 20; i++ {
		body = strings.TrimSpace(body)
		body = strings.TrimSuffix(body, "data:")
		body = strings.TrimSuffix(body, "[DONE]")
	}
	return fmt.Sprintf("data: %s\n", body), true
}
