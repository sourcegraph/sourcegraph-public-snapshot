package clsp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	"github.com/sourcegraph/sourcegraph/internal/env"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog
	_ "github.com/tliron/commonlog/simple"
)

const lsName = "cody"

var (
	version  string = "1.0.0"
	handler  protocol.Handler
	debug, _ = strconv.ParseBool(env.Get("CLSP_DEBUG", strconv.FormatBool(env.InsecureDev), "Enable debugging for CLSP server (app only)"))
)

func Server() error {
	// This increases logging verbosity (optional)
	commonlog.Configure(1, nil)

	handler = protocol.Handler{
		Initialize:        initialize,
		Initialized:       initialized,
		TextDocumentHover: textDocumentHover,
		Shutdown:          shutdown,
		SetTrace:          setTrace,
	}

	server := server.NewServer(&handler, lsName, false)

	return server.RunTCP("127.0.0.1:65444")
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

var tokenizeRe = regexp.MustCompile(`[a-zA-Z0-9\.]+|.`)

func sliceLines(lines []string, start, end int) []string {
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		start = len(lines)
	}
	if end < 0 {
		end = 0
	}
	if end > len(lines) {
		end = len(lines)
	}
	return lines[start:end]
}

func textDocumentHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	// Grab the file contents
	filePath, _ := url.PathUnescape(strings.TrimPrefix(params.TextDocument.URI, "file://"))
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Grab the line the hover request was over.
	lines := strings.Split(string(fileContents), "\n")
	line := lines[params.Position.Line]
	likelyImports := sliceLines(lines, 0, 10)
	numContextLines := 5
	contextLines := sliceLines(lines, int(params.Position.Line)-numContextLines, int(params.Position.Line)+numContextLines)

	// Extract the token that was being hovered over.
	tokens := tokenizeRe.FindAllString(line, -1)
	token := ""
	offset := 0
	for _, tok := range tokens {
		offset += len(tok)
		if offset > int(params.Position.Character) {
			token = tok
			break
		}
	}

	human := `$FILE

Pretend you are an IDE backend, which should respond with Markdown content when a user hovers over code in the above file.

This line of code: $LINE
This token specifically: $TOKEN
Do not suggest that you are just an AI assistant, describe the code and provide documentation for it
Provide a short usage example code block if possible
Prefix your response with "**Cody**"

`
	human = strings.ReplaceAll(human, "$TOKEN", token)
	human = strings.ReplaceAll(human, "$LINE", line)
	human = strings.ReplaceAll(human, "$FILE", "```\n"+strings.Join(likelyImports, "\n")+"\n...\n"+strings.Join(contextLines, "\n")+"\n```\n")

	if debug {
		fmt.Println("textDocument/hover <-", human)
	}
	assistant := ``
	fast := false
	start := time.Now()
	answer, err := codyCompletions(context.Background(), human, assistant, fast)
	if debug {
		fmt.Println("textDocument/hover ->", answer)
		fmt.Println("textDocument/hover finished in", time.Since(start))
	}
	if err != nil {
		return nil, err
	}

	var resp protocol.Hover
	resp.Contents = protocol.MarkupContent{Kind: "markdown", Value: answer}
	return &resp, nil
}
