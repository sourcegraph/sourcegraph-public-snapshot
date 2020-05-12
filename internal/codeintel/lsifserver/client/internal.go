package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
)

func (c *Client) Prune(ctx context.Context) (int64, bool, error) {
	req := &lsifRequest{
		method: "POST",
		path:   "/prune",
	}

	var payload struct {
		ID *int64 `json:"id"`
	}
	if _, err := c.do(ctx, req, &payload); err != nil {
		return 0, false, err
	}

	if payload.ID == nil {
		return 0, false, nil
	}
	return *payload.ID, true, nil
}

func (c *Client) States(ctx context.Context, ids []int) (map[int]string, error) {
	serialized, err := json.Marshal(map[string]interface{}{"ids": ids})
	if err != nil {
		return nil, err
	}

	req := &lsifRequest{
		method: "POST",
		path:   "/uploads",
		body:   ioutil.NopCloser(bytes.NewReader(serialized)),
	}

	var payload struct {
		Value []json.RawMessage `json:"value"`
	}
	if _, err := c.do(ctx, req, &payload); err != nil {
		return nil, err
	}

	states := map[int]string{}
	for _, pair := range payload.Value {
		var key int
		var value string
		payload := []interface{}{&key, &value}
		if err := json.Unmarshal([]byte(pair), &payload); err != nil {
			return nil, err
		}

		states[key] = value
	}

	return states, nil
}
