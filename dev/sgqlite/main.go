package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bndr/gotabulate"
)

const gqlQuery = `
query Execute($query: String!){
  execute(query:$query) {
    columnNames
    rows
  }
}
`

type response struct {
	Data struct {
		Execute struct {
			ColumnNames []string
			Rows        [][]interface{}
		}
	}
	Errors []struct {
		Message string
	}
}

func main() {
	query, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	variables := map[string]interface{}{
		"query": string(query),
	}

	body, err := json.Marshal(map[string]interface{}{
		"query":     gqlQuery,
		"variables": variables,
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://sourcegraph.test:3443/.api/graphql", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token 3ebba4c955055996db2bd110c39dc5471be67ec2"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("non-200")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var unmarshaledResp response
	if err := json.Unmarshal(respBody, &unmarshaledResp); err != nil {
		panic(err)
	}

	if len(unmarshaledResp.Errors) > 0 {
		fmt.Printf("%#v\n", unmarshaledResp.Errors)
		return
	}

	t := gotabulate.Create(unmarshaledResp.Data.Execute.Rows)
	t.SetHeaders(unmarshaledResp.Data.Execute.ColumnNames)
	fmt.Println(t.Render("grid"))
}
