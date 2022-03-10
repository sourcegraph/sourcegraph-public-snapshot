package squirrel

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/grafana/regexp"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

//go:embed queries
var queriesFs embed.FS

//go:embed language-file-extensions.json
var languageFileExtensionsJson string

// Mapping from langauge name to file extensions.
var langToExts = func() map[string][]string {
	var m map[string][]string
	err := json.Unmarshal([]byte(languageFileExtensionsJson), &m)
	if err != nil {
		panic(err)
	}
	return m
}()

// Mapping from file extension to language name.
var extToLang = func() map[string]string {
	m := map[string]string{}
	for lang, exts := range langToExts {
		for _, ext := range exts {
			if _, ok := m[ext]; ok {
				panic(fmt.Sprintf("duplicate file extension %s", ext))
			}
			m[ext] = lang
		}
	}
	return m
}()

// Info about a language.
type LangSpec struct {
	nvimQueryDir string
	language     *sitter.Language
	commentStyle CommentStyle
}

// Info about comments in a language.
type CommentStyle struct {
	placedBelow   bool
	ignoreRegex   *regexp.Regexp
	stripRegex    *regexp.Regexp
	skipNodeTypes []string
	nodeTypes     []string
	codeFenceName string
}

// Mapping from language name to language specification.
var langToLangSpec = map[string]LangSpec{
	"cpp": {
		nvimQueryDir: "cpp",
		language:     cpp.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"csharp": {
		nvimQueryDir: "c_sharp",
		language:     csharp.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"css": {
		nvimQueryDir: "css",
		language:     css.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"dockerfile": {
		nvimQueryDir: "dockerfile",
		language:     dockerfile.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"elm": {
		nvimQueryDir: "elm",
		language:     elm.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"go": {
		nvimQueryDir: "go",
		language:     golang.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^//`),
			codeFenceName: "go",
		}, // TODO
	},
	"hcl": {
		nvimQueryDir: "hcl",
		language:     hcl.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"html": {
		nvimQueryDir: "html",
		language:     html.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"java": {
		nvimQueryDir: "java",
		language:     java.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"line_comment", "block_comment"},
			stripRegex:    regexp.MustCompile(`(^//|^\s*\*|^/\*\*|\*/$)`),
			ignoreRegex:   regexp.MustCompile(`^\s*(/\*\*|\*/)\s*$`),
			codeFenceName: "java",
			skipNodeTypes: []string{"modifiers"},
		}, // TODO
	},
	"javascript": {
		nvimQueryDir: "javascript",
		language:     javascript.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"lua": {
		nvimQueryDir: "lua",
		language:     lua.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"ocaml": {
		nvimQueryDir: "ocaml",
		language:     ocaml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"php": {
		nvimQueryDir: "php",
		language:     php.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"python": {
		nvimQueryDir: "python",
		language:     python.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"ruby": {
		nvimQueryDir: "ruby",
		language:     ruby.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"rust": {
		nvimQueryDir: "rust",
		language:     rust.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"scala": {
		nvimQueryDir: "scala",
		language:     scala.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"shell": {
		nvimQueryDir: "bash",
		language:     bash.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"svelte": {
		nvimQueryDir: "svelte",
		language:     svelte.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"toml": {
		nvimQueryDir: "toml",
		language:     toml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"typescript": {
		nvimQueryDir: "typescript",
		language:     typescript.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
	"yaml": {
		nvimQueryDir: "yaml",
		language:     yaml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
	},
}
