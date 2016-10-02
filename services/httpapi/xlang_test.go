package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
)

func TestXLang(t *testing.T) {
	c, _ := newTest()

	orig := xlangCreateConnection
	defer func() {
		xlangCreateConnection = orig
	}()
	var xc xlangTestClient
	xlangCreateConnection = func() (xlangClient, error) { return &xc, nil }

	postJSON := func(lspMethod string, h http.Header, reqBodyJSON string, respBody interface{}) error {
		req, err := http.NewRequest("POST", "/xlang/"+lspMethod, strings.NewReader(reqBodyJSON))
		if err != nil {
			return err
		}
		req.Header.Set("content-type", "application/json; charset=utf-8")
		for k, v := range h {
			req.Header[k] = v
		}
		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP error status %d", resp.StatusCode)
		}
		if respBody != nil {
			if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
				return err
			}
		}
		return nil
	}

	if err := postJSON("someMethod", nil, `[{"id":0,"method":"initialize","params":{"rootPath":"/"}},{"id":1,"method":"someMethod","params":{}},{"id":2,"method":"shutdown"},{"id":3,"method":"exit"}]`, nil); err != nil {
		t.Fatal(err)
	}
	if want := []string{"initialize", "someMethod", "shutdown", "exit"}; !reflect.DeepEqual(xc.methodsCalled, want) {
		t.Errorf("got methods called == %v, want %v", xc.methodsCalled, want)
	}
}

type xlangTestClient struct{ methodsCalled []string }

func (c *xlangTestClient) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error {
	c.methodsCalled = append(c.methodsCalled, method)
	return nil
}

func (c *xlangTestClient) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	c.methodsCalled = append(c.methodsCalled, method)
	return nil
}

func (c *xlangTestClient) Close() error { return nil }
