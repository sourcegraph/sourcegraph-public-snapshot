package coverage

import (
	"context"
	"fmt"
	"net"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// Client is a coverage client type, which is effectively either an LSP client
// or a langp client.
type Client interface {
	Definition(ctx context.Context, p *langp.Position) (*langp.Range, error)
	Hover(ctx context.Context, p *langp.Position) (*langp.Hover, error)
	LocalRefs(ctx context.Context, p *langp.Position) (*langp.RefLocations, error)
	Close() error
}

type noopCloser struct {
	*langp.Client
}

func (c *noopCloser) Close() error {
	return nil
}

func LangpClient(addr string) (Client, error) {
	c, err := langp.NewClient(addr)
	if err != nil {
		return nil, err
	}
	return &noopCloser{Client: c}, nil
}

func LSPClient(addr, rootPath string, printCaps bool) (Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &lspClient{
		c: &betterClient{
			c: jsonrpc2.NewClient(conn),
		},
		rootPath: rootPath,
	}
	if err := c.checkServerCaps(printCaps); err != nil {
		return nil, err
	}

	return c, nil
}

type lspClient struct {
	c        *betterClient
	rootPath string
}

func (c *lspClient) Definition(ctx context.Context, p *langp.Position) (*langp.Range, error) {
	var (
		initResp lsp.InitializeResult
		locResp  lsp.Location
	)
	if err := c.c.RequestBatchAndWaitForAllResponses(
		betterRequest{
			Method: "initialize",
			Params: &lsp.InitializeParams{
				RootPath: c.rootPath,
			},
			Results: &initResp,
		},
		betterRequest{
			Method:  "textDocument/definition",
			Params:  p.LSP(),
			Results: &locResp, // TODO: this can return multiple locations.. BetterClient doesn't handle that.
		},
		betterRequest{Method: "shutdown"},
	); err != nil {
		return nil, err
	}
	return &langp.Range{
		Repo:           p.Repo,
		Commit:         p.Commit,
		File:           p.File,
		StartLine:      locResp.Range.Start.Line,
		StartCharacter: locResp.Range.Start.Character,
		EndLine:        locResp.Range.End.Line,
		EndCharacter:   locResp.Range.End.Character,
	}, nil
}

func (c *lspClient) Hover(ctx context.Context, p *langp.Position) (*langp.Hover, error) {
	var (
		initResp  lsp.InitializeResult
		hoverResp lsp.Hover
	)
	if err := c.c.RequestBatchAndWaitForAllResponses(
		betterRequest{
			Method: "initialize",
			Params: &lsp.InitializeParams{
				RootPath: c.rootPath,
			},
			Results: &initResp,
		},
		betterRequest{
			Method:  "textDocument/hover",
			Params:  p.LSP(),
			Results: &hoverResp, // TODO: this can return multiple locations.. BetterClient doesn't handle that.
		},
		betterRequest{Method: "shutdown"},
	); err != nil {
		return nil, err
	}
	return langp.HoverFromLSP(&hoverResp), nil
}

func (c *lspClient) LocalRefs(ctx context.Context, p *langp.Position) (*langp.RefLocations, error) {
	var (
		initResp lsp.InitializeResult
		refsResp []*lsp.Location
	)
	if err := c.c.RequestBatchAndWaitForAllResponses(
		betterRequest{
			Method: "initialize",
			Params: &lsp.InitializeParams{
				RootPath: c.rootPath,
			},
			Results: &initResp,
		},
		betterRequest{
			Method: "textDocument/references",
			Params: &lsp.ReferenceParams{
				TextDocumentPositionParams: *p.LSP(),
				Context: lsp.ReferenceContext{
					IncludeDeclaration: false, // for posterity
				},
			},
			Results: &refsResp,
		},
		betterRequest{Method: "shutdown"},
	); err != nil {
		return nil, err
	}
	panic("not fully implemented (translate refsResp to langp.LocalRefs")
}

func (c *lspClient) Close() error {
	return c.c.Close()
}

func (c *lspClient) checkServerCaps(printCaps bool) error {
	var (
		initResp lsp.InitializeResult
	)
	if err := c.c.RequestBatchAndWaitForAllResponses(
		betterRequest{
			Method:  "initialize",
			Params:  &lsp.InitializeParams{},
			Results: &initResp,
		},
		betterRequest{Method: "shutdown"},
	); err != nil {
		return err
	}

	caps := initResp.Capabilities

	if printCaps {
		fmt.Println("")
		fmt.Println("Server Capabilities:")
		fmt.Println("	CodeActionProvider:", caps.CodeActionProvider)
		fmt.Println("	CodeLensProvider:", caps.CodeLensProvider)
		fmt.Println("	CompletionProvider:", caps.CompletionProvider)
		fmt.Println("	DefinitionProvider:", caps.DefinitionProvider)
		fmt.Println("	DocumentFormattingProvider:", caps.DocumentFormattingProvider)
		fmt.Println("	DocumentHighlightProvider:", caps.DocumentHighlightProvider)
		fmt.Println("	DocumentOnTypeFormattingProvider:", caps.DocumentOnTypeFormattingProvider)
		fmt.Println("	DocumentRangeFormattingProvider:", caps.DocumentRangeFormattingProvider)
		fmt.Println("	DocumentSymbolProvider:", caps.DocumentSymbolProvider)
		fmt.Println("	HoverProvider:", caps.HoverProvider)
		fmt.Println("	ReferencesProvider:", caps.ReferencesProvider)
		fmt.Println("	RenameProvider:", caps.RenameProvider)
		fmt.Println("	SignatureHelpProvider:", caps.SignatureHelpProvider)
		fmt.Println("	TextDocumentSync:", caps.TextDocumentSync)
		fmt.Println("	WorkspaceSymbolProvider:", caps.WorkspaceSymbolProvider)
		fmt.Println("")
		os.Exit(0)
	}

	if !caps.DefinitionProvider {
		return fmt.Errorf("Server does not have Definition capability")
	}
	if !caps.HoverProvider {
		return fmt.Errorf("Server does not have Hover capability")
	}
	if !caps.ReferencesProvider {
		return fmt.Errorf("Server does not have References capability")
	}
	return nil
}
