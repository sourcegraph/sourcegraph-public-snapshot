package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"

	"github.com/sourcegraph/srclib-json-tokenizer/sgjson"
)

func init() {
	_, err := parser.AddCommand("graph",
		"graph a JSON file",
		"graph a JSON file, producing all refs", &graphCmd)
	if err != nil {
		log.Fatal(err)
	}
}

type GraphCmd struct{}

var graphCmd GraphCmd

func (c *GraphCmd) Execute(args []string) error {
	inputBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var unit unit.SourceUnit
	if err = json.Unmarshal(inputBytes, &unit); err != nil {
		return err
	}
	out, err := doGraph(unit)
	if err != nil {
		return err
	}

	outBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(outBytes); err != nil {
		return err
	}
	return nil
}

func doGraph(u unit.SourceUnit) (*graph.Output, error) {
	out := &graph.Output{}
	for _, file := range u.Files {
		fileBytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("Unable to load file: %s", err)
		}
		tokenInfos, err := TokenizeJSON(bytes.NewReader(fileBytes))
		if err != nil {
			return nil, fmt.Errorf("Unable to tokenize JSON in file %s, error: %s", file, err)
		}
		for _, info := range tokenInfos {
			tokenString := string(fileBytes[info.Start:info.Endp])
			out.Refs = append(out.Refs, &graph.Ref{
				Start:       uint32(info.Start),
				End:         uint32(info.Endp),
				File:        file,
				Unit:        u.Name,
				UnitType:    "json",
				DefPath:     defPath(file, tokenString, info),
				DefUnit:     u.Name,
				DefUnitType: "placeholder-type"})
		}
	}
	return out, nil
}

// TokenizeJSON - given a reader "r" with a JSON object inside it, returns a slice of all
// non-delimiter TokenInfos in the JSON
func TokenizeJSON(r io.Reader) ([]sgjson.TokenInfo, error) {
	dec := sgjson.NewDecoder(r)
	dec.UseNumber()
	tokens, err := dec.Tokenize()
	if err != nil {
		return nil, err
	}
	var out []sgjson.TokenInfo
	for _, t := range tokens {
		switch t.Token.(type) {
		case sgjson.Delim:
			continue
		case string:
			//remove quotations
			t.Start++
			t.Endp--
		}
		out = append(out, t)
	}
	return out, nil
}

// looks like: ...[relativePath]/keyPath[0]/keyPath[1]/.../tokenString/token.(Type)/(isKey)
func defPath(filePath, tokenString string, t sgjson.TokenInfo) string {
	fileName := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	elems := []string{fileName}
	for _, key := range t.KeyPath {
		elems = append(elems, key)
	}

	// in case there is a slash in the token itself, so that it doesn't conflict with paths
	escapedString := strings.Replace(tokenString, `/`, `\\/`, -1)
	elems = append(elems, escapedString)
	elems = append(elems, fmt.Sprintf("%T", t.Token))
	if t.IsKey {
		elems = append(elems, "is_key")
	} else {
		elems = append(elems, "not_key")
	}
	for i, elem := range elems {
		elems[i] = strings.TrimSpace(elem)
	}
	return strings.Join(elems, "/")
}
