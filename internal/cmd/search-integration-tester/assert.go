package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/google/go-cmp/cmp"
)

func assertGolden(path string, got GQLResult) error {
	wantString, err := ioutil.ReadFile(path)
	if err != nil {
		// Doesn't exist, set empty to empty object to see the diff.
		wantString = []byte("{}")
	}
	var want interface{}
	err = json.Unmarshal(wantString, &want)
	if err != nil {
		return err
	}
	if diff := cmp.Diff(want, got); diff != "" {
		return errors.New(diff)
	}
	return nil
}

func assertUpdate(path string, got GQLResult) error {
	gotBytes, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("could not marshal response %s", string(gotBytes)))
	}
	if update || updateAll {
		if err := ioutil.WriteFile(path, gotBytes, 0640); err != nil {
			return fmt.Errorf("failed to update golden file %q: %s", path, err)
		}
	}
	return nil
}
