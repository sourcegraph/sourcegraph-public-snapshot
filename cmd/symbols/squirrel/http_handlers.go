package squirrel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/inconshreveable/log15"
	sitter "github.com/smacker/go-tree-sitter"

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Responds to /localCodeIntel
func LocalCodeIntelHandler(readFile ReadFileFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the args from the request body.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log15.Error("failed to read request body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var args types.RepoCommitPath
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&args); err != nil {
			log15.Error("failed to decode request body", "err", err, "body", string(body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		squirrel := New(readFile, nil)
		defer squirrel.Close()

		// Compute the local code intel payload.
		payload, err := squirrel.localCodeIntel(r.Context(), args)
		if payload != nil && os.Getenv("SQUIRREL_DEBUG") == "true" {
			debugStringBuilder := &strings.Builder{}
			fmt.Fprintln(debugStringBuilder, "👉 /localCodeIntel repo:", args.Repo, "commit:", args.Commit, "path:", args.Path)
			contents, err := readFile(r.Context(), args)
			if err != nil {
				log15.Error("failed to read file from gitserver", "err", err)
			} else {
				prettyPrintLocalCodeIntelPayload(debugStringBuilder, *payload, string(contents))
				fmt.Fprintln(debugStringBuilder, "✅ /localCodeIntel repo:", args.Repo, "commit:", args.Commit, "path:", args.Path)

				fmt.Println(" ")
				fmt.Println(bracket(debugStringBuilder.String()))
				fmt.Println(" ")
			}
		}
		if err != nil {
			_ = json.NewEncoder(w).Encode(nil)

			// Log the error if it's not an unrecognized file extension or unsupported language error.
			if !errors.Is(err, unrecognizedFileExtensionError) && !errors.Is(err, unsupportedLanguageError) {
				log15.Error("failed to generate local code intel payload", "err", err)
			}

			return
		}

		// Write the response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			log15.Error("failed to write response: %s", "error", err)
			http.Error(w, fmt.Sprintf("failed to generate local code intel payload: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

// Responds to /symbolInfo
func NewSymbolInfoHandler(symbolSearch symbolsTypes.SearchFunc, readFile ReadFileFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the args from the request body.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log15.Error("failed to read request body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var args types.RepoCommitPathPoint
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&args); err != nil {
			log15.Error("failed to decode request body", "err", err, "body", string(body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Find the symbol.
		squirrel := New(readFile, symbolSearch)
		defer squirrel.Close()
		result, err := squirrel.symbolInfo(r.Context(), args)
		if os.Getenv("SQUIRREL_DEBUG") == "true" {
			debugStringBuilder := &strings.Builder{}
			fmt.Fprintln(debugStringBuilder, "👉 /symbolInfo repo:", args.Repo, "commit:", args.Commit, "path:", args.Path, "row:", args.Row, "column:", args.Column)
			squirrel.breadcrumbs.pretty(debugStringBuilder, readFile)
			if result == nil {
				fmt.Fprintln(debugStringBuilder, "❌ no definition found")
			} else {
				fmt.Fprintln(debugStringBuilder, "✅ /symbolInfo", *result)
			}

			fmt.Println(" ")
			fmt.Println(bracket(debugStringBuilder.String()))
			fmt.Println(" ")
		}
		if err != nil {
			_ = json.NewEncoder(w).Encode(nil)
			log15.Error("failed to get definition", "err", err)
			return
		}

		// Write the response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log15.Error("failed to write response: %s", "error", err)
			http.Error(w, fmt.Sprintf("failed to get definition: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

// Response to /debugLocalCodeIntel.
func DebugLocalCodeIntelHandler(w http.ResponseWriter, r *http.Request) {
	// Read ?ext=<ext> from the request.
	ext := r.URL.Query().Get("ext")
	if ext == "" {
		http.Error(w, "missing ?ext=<ext> query parameter", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	path := types.RepoCommitPath{
		Repo:   "foo",
		Commit: "bar",
		Path:   "example." + ext,
	}

	fileToRead := "/tmp/squirrel-example.txt"
	readFile := func(ctx context.Context, args types.RepoCommitPath) ([]byte, error) {
		return os.ReadFile("/tmp/squirrel-example.txt")
	}

	squirrel := New(readFile, nil)
	defer squirrel.Close()

	rangeToSymbolIx := map[types.Range]int{}
	symbolIxToColor := map[int]string{}
	payload, err := squirrel.localCodeIntel(r.Context(), path)
	if err != nil {
		fmt.Fprintf(w, "failed to generate local code intel payload: %s\n\n", err)
	} else {
		for ix := range payload.Symbols {
			nonRed := []int{100, 120, 140, 160, 180, 200, 220, 240, 260}
			symbolIxToColor[ix] = fmt.Sprintf("hsla(%d, 100%%, 50%%, 0.5)", sample(nonRed))
		}

		for ix, symbol := range payload.Symbols {
			rangeToSymbolIx[symbol.Def] = ix
			for _, ref := range symbol.Refs {
				rangeToSymbolIx[ref] = ix
			}
		}
	}

	node, err := squirrel.parse(r.Context(), path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `
		<style>
			span:hover {
				outline: 2px solid red;
			}
		</style>
	`)

	fmt.Fprintf(w, "<h3>Parsing as %s from file on disk %s</h3>\n", ext, fileToRead)

	var nodeToHtml func(*sitter.Node, string) string
	nodeToHtml = func(n *sitter.Node, stack string) string {

		thisStack := stack + ">" + n.Type()

		// Pick color
		color := ""
		if n.Type() == "ERROR" {
			color = "hsla(0, 100%, 50%, 0.2)"
		} else if ix, ok := rangeToSymbolIx[nodeToRange(n)]; ok {
			if c, ok := symbolIxToColor[ix]; ok {
				color = c
			}
		} else {
			color = "hsla(0, 0%, 0%, 0.0)"
		}

		// Tooltip
		title := fmt.Sprintf("%s %d:%d-%d:%d", thisStack, n.StartPoint().Row, n.StartPoint().Column, n.EndPoint().Row, n.EndPoint().Column)

		if n.ChildCount() == 0 {

			// Base case: no children

			// Render
			return fmt.Sprintf(
				`<span style="background-color: %s", title="%s">%s</span>`,
				color,
				title,
				html.EscapeString(string(node.Contents[n.StartByte():n.EndByte()])),
			)
		} else {

			// Recursive case: with children

			// Concatenate children
			b := n.StartByte()
			inner := &strings.Builder{}
			for i := 0; i < int(n.ChildCount()); i++ {
				inner.WriteString(html.EscapeString(string(node.Contents[b:n.Child(i).StartByte()])))
				inner.WriteString(nodeToHtml(n.Child(i), thisStack))
				b = n.Child(i).EndByte()
			}
			inner.WriteString(html.EscapeString(string(node.Contents[b:n.EndByte()])))

			// Render
			return fmt.Sprintf(
				`<span style="background-color: %s", title="%s">%s</span>`,
				color,
				title,
				inner.String(),
			)
		}
	}

	fmt.Fprint(w, "<pre>"+nodeToHtml(node.Node, "")+"</pre>")
}

func sample[T any](xs []T) T {
	return xs[rand.Intn(len(xs))]
}
