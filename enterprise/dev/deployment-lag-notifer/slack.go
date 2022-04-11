package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SlackClient struct {
	WebhookURL string

	HttpClient *http.Client
}

func NewSlackClient(url string) *SlackClient {
	hc := http.Client{}

	c := SlackClient{
		WebhookURL: url,
		HttpClient: &hc,
	}

	return &c
}

func (s *SlackClient) PostMessage(b bytes.Buffer) error {

	type slackRequest struct {
		Text     string `json:"text"`
		Markdown bool   `json:"mrkdwn"`
	}

	payload, err := json.Marshal(slackRequest{Text: b.String(), Markdown: true})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	return nil
}
