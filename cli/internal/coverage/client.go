package coverage

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// Client is a coverage client type, which is effectively an LSP client.
type Client interface {
	Definition(ctx context.Context, p *Position) (*Range, error)
	Hover(ctx context.Context, p *Position) (*Hover, error)
	LocalRefs(ctx context.Context, p *Position) (*RefLocations, error)
	Close() error
}

func LSPClient(addr, rootPath string, printCaps bool) (Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	c := &lspClient{
		c:        jsonrpc2.NewConn(ctx, conn, nil),
		rootPath: rootPath,
	}
	if err := c.checkServerCaps(ctx, printCaps); err != nil {
		return nil, err
	}

	return c, nil
}

type lspClient struct {
	c        *jsonrpc2.Conn
	rootPath string
}

func (c *lspClient) Definition(ctx context.Context, p *Position) (*Range, error) {
	if err := c.c.Call(ctx, "initialize", &lsp.InitializeParams{RootPath: c.rootPath}, nil); err != nil {
		return nil, err
	}
	var locResp []lsp.Location
	if err := c.c.Call(ctx, "textDocument/definition", p.LSP(), &locResp); err != nil {
		return nil, err
	}
	if err := c.c.Call(ctx, "shutdown", nil, nil); err != nil {
		return nil, err
	}
	if len(locResp) != 1 {
		return nil, fmt.Errorf("could not find definition")
	}
	return &Range{
		Repo:           p.Repo,
		Commit:         p.Commit,
		File:           p.File,
		StartLine:      locResp[0].Range.Start.Line,
		StartCharacter: locResp[0].Range.Start.Character,
		EndLine:        locResp[0].Range.End.Line,
		EndCharacter:   locResp[0].Range.End.Character,
	}, nil
}

func (c *lspClient) Hover(ctx context.Context, p *Position) (*Hover, error) {
	if err := c.c.Call(ctx, "initialize", &lsp.InitializeParams{RootPath: c.rootPath}, nil); err != nil {
		return nil, err
	}
	var hoverResp lsp.Hover
	if err := c.c.Call(ctx, "textDocument/hover", p.LSP(), &hoverResp); err != nil {
		return nil, err
	}
	if err := c.c.Call(ctx, "shutdown", nil, nil); err != nil {
		return nil, err
	}
	return HoverFromLSP(&hoverResp), nil
}

func (c *lspClient) LocalRefs(ctx context.Context, p *Position) (*RefLocations, error) {
	if err := c.c.Call(ctx, "initialize", &lsp.InitializeParams{RootPath: c.rootPath}, nil); err != nil {
		return nil, err
	}
	rp := lsp.ReferenceParams{
		TextDocumentPositionParams: *p.LSP(),
		Context: lsp.ReferenceContext{
			IncludeDeclaration: false, // for posterity
		},
	}
	var refsResp []*lsp.Location
	if err := c.c.Call(ctx, "textDocument/references", rp, &refsResp); err != nil {
		return nil, err
	}
	if err := c.c.Call(ctx, "shutdown", nil, nil); err != nil {
		return nil, err
	}
	panic("not fully implemented (translate refsResp to LocalRefs")
}

func (c *lspClient) Close() error {
	return c.c.Close()
}

func (c *lspClient) checkServerCaps(ctx context.Context, printCaps bool) error {
	var initResp lsp.InitializeResult
	if err := c.c.Call(ctx, "initialize", &lsp.InitializeParams{RootPath: c.rootPath}, &initResp); err != nil {
		return err
	}

	if err := c.c.Call(ctx, "shutdown", nil, nil); err != nil {
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
