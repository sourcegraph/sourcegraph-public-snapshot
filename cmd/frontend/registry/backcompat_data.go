package registry

// backcompatLanguageServers maps language keys to static information about the language
// server. For each entry, the siteConfig.Address must match the port specified
// in the corresponding Dockerfile.
var backcompatLanguageServers = map[string]*struct {
	// DisplayName is the display name of the language. e.g. "PHP" and
	// "TypeScript", rather than language keys which are always lowercase "php"
	// and "typescript".
	DisplayName string

	// HomepageURL is the URL to the language server's homepage, or an empty
	// string if there is none.
	HomepageURL string

	// IssuesURL is the URL to the language server's open/known issues, or an
	// empty string if there is none.
	IssuesURL string

	// DocsURL is the URL to the language server's documentation, or an empty
	// string if there is none.
	DocsURL string

	// Experimental indicates that a language server may perform arbitrary code
	// execution, may have limited functionality, etc.
	Experimental bool
}{
	"go": {
		DisplayName:  "Go",
		HomepageURL:  "https://github.com/sourcegraph/go-langserver",
		IssuesURL:    "https://github.com/sourcegraph/go-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/go-langserver/blob/master/README.md",
		Experimental: false,
	},
	"typescript": {
		DisplayName:  "TypeScript",
		HomepageURL:  "https://github.com/sourcegraph/javascript-typescript-langserver",
		IssuesURL:    "https://github.com/sourcegraph/javascript-typescript-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/javascript-typescript-langserver/blob/master/README.md",
		Experimental: false,
	},
	"javascript": {
		DisplayName:  "JavaScript",
		HomepageURL:  "https://github.com/sourcegraph/javascript-typescript-langserver",
		IssuesURL:    "https://github.com/sourcegraph/javascript-typescript-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/javascript-typescript-langserver/blob/master/README.md",
		Experimental: false,
	},
	"python": {
		DisplayName:  "Python",
		HomepageURL:  "https://github.com/sourcegraph/python-langserver",
		IssuesURL:    "https://github.com/sourcegraph/python-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/python-langserver/blob/master/README.md",
		Experimental: false,
	},
	"java": {
		DisplayName:  "Java",
		HomepageURL:  "https://github.com/sourcegraph/java-langserver-docs",
		IssuesURL:    "https://github.com/sourcegraph/java-langserver-docs/issues",
		DocsURL:      "https://github.com/sourcegraph/java-langserver-docs/blob/master/README.md",
		Experimental: false,
	},
	"php": {
		DisplayName:  "PHP",
		HomepageURL:  "https://github.com/felixfbecker/php-language-server",
		IssuesURL:    "https://github.com/felixfbecker/php-language-server/issues",
		DocsURL:      "https://github.com/felixfbecker/php-language-server/blob/master/README.md",
		Experimental: false,
	},
	"bash": {
		DisplayName:  "Bash",
		HomepageURL:  "https://github.com/mads-hartmann/bash-language-server",
		IssuesURL:    "https://github.com/mads-hartmann/bash-language-server/issues",
		DocsURL:      "https://github.com/mads-hartmann/bash-language-server/blob/master/README.md",
		Experimental: true,
	},
	"clojure": {
		DisplayName:  "Clojure",
		HomepageURL:  "https://github.com/snoe/clojure-lsp",
		IssuesURL:    "https://github.com/snoe/clojure-lsp/issues",
		DocsURL:      "https://github.com/snoe/clojure-lsp/blob/master/README.md",
		Experimental: true,
	},
	"cpp": {
		DisplayName:  "C++",
		HomepageURL:  "https://github.com/Chilledheart/vim-clangd",
		IssuesURL:    "https://github.com/Chilledheart/vim-clangd/issues",
		DocsURL:      "https://github.com/Chilledheart/vim-clangd/blob/master/README.md",
		Experimental: true,
	},
	"cs": {
		DisplayName:  "C#",
		HomepageURL:  "https://github.com/OmniSharp/omnisharp-node-client",
		IssuesURL:    "https://github.com/OmniSharp/omnisharp-node-client/issues",
		DocsURL:      "https://github.com/OmniSharp/omnisharp-node-client/blob/master/readme.md",
		Experimental: true,
	},
	"css": {
		DisplayName:  "CSS",
		HomepageURL:  "https://github.com/vscode-langservers/vscode-css-languageserver-bin",
		IssuesURL:    "https://github.com/vscode-langservers/vscode-css-languageserver-bin/issues",
		DocsURL:      "https://github.com/vscode-langservers/vscode-css-languageserver-bin/blob/master/README.md",
		Experimental: true,
	},
	"dockerfile": {
		DisplayName:  "Dockerfile",
		HomepageURL:  "https://github.com/rcjsuen/dockerfile-language-server-nodejs",
		IssuesURL:    "https://github.com/rcjsuen/dockerfile-language-server-nodejs/issues",
		DocsURL:      "https://github.com/rcjsuen/dockerfile-language-server-nodejs/blob/master/README.md",
		Experimental: true,
	},
	"elixir": {
		DisplayName:  "Elixir",
		HomepageURL:  "https://github.com/JakeBecker/elixir-ls",
		IssuesURL:    "https://github.com/JakeBecker/elixir-ls/issues",
		DocsURL:      "https://github.com/JakeBecker/elixir-ls/blob/master/README.md",
		Experimental: true,
	},
	"haskell": {
		DisplayName:  "Haskell",
		HomepageURL:  "https://github.com/haskell/haskell-ide-engine",
		IssuesURL:    "https://github.com/haskell/haskell-ide-engine/issues",
		DocsURL:      "https://github.com/haskell/haskell-ide-engine/blob/master/README.md",
		Experimental: true,
	},
	"html": {
		DisplayName:  "HTML",
		HomepageURL:  "https://github.com/vscode-langservers/vscode-html-languageserver-bin",
		IssuesURL:    "https://github.com/vscode-langservers/vscode-html-languageserver-bin/issues",
		DocsURL:      "https://github.com/vscode-langservers/vscode-html-languageserver-bin/blob/master/README.md",
		Experimental: true,
	},
	"lua": {
		DisplayName:  "Lua",
		HomepageURL:  "https://github.com/Alloyed/lua-lsp",
		IssuesURL:    "https://github.com/Alloyed/lua-lsp/issues",
		DocsURL:      "https://github.com/Alloyed/lua-lsp/blob/master/readme.md",
		Experimental: true,
	},
	"ocaml": {
		DisplayName:  "OCaml",
		HomepageURL:  "https://github.com/freebroccolo/ocaml-language-server",
		IssuesURL:    "https://github.com/freebroccolo/ocaml-language-server/issues",
		DocsURL:      "https://github.com/freebroccolo/ocaml-language-server/blob/master/README.md",
		Experimental: true,
	},
	"r": {
		DisplayName:  "R",
		HomepageURL:  "https://github.com/REditorSupport/languageserver",
		IssuesURL:    "https://github.com/REditorSupport/languageserver/issues",
		DocsURL:      "https://github.com/REditorSupport/languageserver/blob/master/README.md",
		Experimental: true,
	},
	"ruby": {
		DisplayName:  "Ruby",
		HomepageURL:  "https://github.com/castwide/solargraph",
		IssuesURL:    "https://github.com/castwide/solargraph/issues",
		DocsURL:      "https://github.com/castwide/solargraph/blob/master/README.md",
		Experimental: true,
	},
	"rust": {
		DisplayName:  "Rust",
		HomepageURL:  "https://github.com/rust-lang-nursery/rls",
		IssuesURL:    "https://github.com/rust-lang-nursery/rls/issues",
		DocsURL:      "https://github.com/rust-lang-nursery/rls/blob/master/README.md",
		Experimental: true,
	},
}
