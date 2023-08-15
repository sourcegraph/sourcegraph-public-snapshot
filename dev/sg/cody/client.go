package cody

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SourcegraphClient struct {
	endpoint string
	token    string
	client   *http.Client
}

type CurrentUserQuery struct {
	Query string `json:"query"`
}

type CurrentUserResponse struct {
	Data struct {
		Username string `json:"username"`
	} `json:"currentUser"`
}

func NewSourcegraphClient(sgURL string, token string) (*SourcegraphClient, error) {
	if !strings.HasSuffix(sgURL, ".api/graphql") {
		api, err := url.JoinPath(sgURL, ".api/graphql")
		if err != nil {
			return nil, err
		}
		sgURL = api
	}

	println(sgURL)
	return &SourcegraphClient{
		endpoint: sgURL,
		token:    token,
		client:   http.DefaultClient,
	}, nil
}

func (s *SourcegraphClient) GetCurrentUser() (string, error) {

	data, err := s.graphqlQuery(CurrentUserQuery{
		Query: `query { currentUser { username } }`,
	})
	if err != nil {
		return "", err
	}

	var u CurrentUserResponse
	if err := json.Unmarshal(data, &u); err != nil {
		return "", err
	}

	return u.Data.Username, nil
}

func (s *SourcegraphClient) graphqlQuery(request any) ([]byte, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "token "+s.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	raw := bytes.NewBuffer(nil)
	if _, err := io.Copy(raw, resp.Body); err != nil {
		return nil, err
	}

	base := map[string]json.RawMessage{}
	err = json.Unmarshal(raw.Bytes(), &base)
	if err != nil {
		return nil, err
	}

	if data, ok := base["data"]; !ok {
		return nil, errors.Newf("missing data entry: %v", raw.String())
	} else {
		return data, err
	}
}
