package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveUserKeys(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	e := json.NewEncoder(w)

	currentUser := handlerutil.UserFromRequest(r)
	if currentUser == nil {
		return fmt.Errorf("user not found")
	}

	// Handle adding a key
	if r.Method == "POST" {
		var data = struct {
			Key, Name string
		}{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&data)
		if err != nil {
			return err
		}

		key := sourcegraph.SSHPublicKey{
			Key:  []byte(data.Key),
			Name: data.Name,
		}

		_, err = apiclient.UserKeys.AddKey(ctx, &key)
		if err != nil {
			return err
		}
	}

	// Handle deleting a key
	if r.Method == "DELETE" {
		id := mux.Vars(r)["id"]
		log.Printf("%#v", id)
	}

	// Then return the current key list
	keys, err := apiclient.UserKeys.ListKeys(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	output := make([]payloads.UserKeysResult, len(keys.SSHKeys))
	for x, key := range keys.SSHKeys {
		output[x].Key = string(key.Key)
		output[x].Name = key.Name
		output[x].Id = int(key.Id)
	}

	return e.Encode(&struct {
		Results []payloads.UserKeysResult
	}{
		Results: output,
	})
}
