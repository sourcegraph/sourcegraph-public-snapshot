package stackimpact

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"
)

type APIRequest struct {
	agent *Agent
}

func newAPIRequest(agent *Agent) *APIRequest {
	ar := &APIRequest{
		agent: agent,
	}

	return ar
}

func (ar *APIRequest) post(endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"runtime_type":    "go",
		"runtime_version": runtime.Version(),
		"agent_version":   AgentVersion,
		"app_name":        ar.agent.AppName,
		"host_name":       ar.agent.HostName,
		"run_id":          ar.agent.runId,
		"run_ts":          ar.agent.runTs,
		"sent_at":         time.Now().Unix(),
		"payload":         payload,
	}

	reqBodyJson, _ := json.Marshal(reqBody)

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(reqBodyJson)
	w.Close()

	url := ar.agent.DashboardAddress + "/agent/v1/" + endpoint
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(ar.agent.AgentKey, "")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	ar.agent.log("Posting API request to %v", url)

	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	resBodyJson, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Received %v: %v", res.StatusCode, string(resBodyJson)))
	} else {
		var resBody map[string]interface{}
		if err := json.Unmarshal(resBodyJson, &resBody); err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot parse response body %v", string(resBodyJson)))
		} else {
			return resBody, nil
		}
	}
}
