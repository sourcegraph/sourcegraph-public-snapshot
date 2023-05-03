package parser

import (
	"path/filepath"
	"strings"
)

// Ported from client/shared/src/languages.ts

func GetLanguageFromPath(path string) string {
	fileName := filepath.Base(path)
	extension := getPathExtension(path)

	exact := getLanguageFromExactFilename(fileName)
	if exact != "" {
		return exact
	}

	ext := getLanguageFromExtension(extension)
	if ext != "" {
		return ext
	}

	return "plaintext"
}

// getLanguageFromExactFilename returns the LSP mode for the
// provided file name (e.g. "dockerfile")
//
// Cherry picked from https://github.com/github/linguist/blob/master/lib/linguist/languages.yml
func getLanguageFromExactFilename(fileName string) string {
	switch strings.ToLower(fileName) {
	case "dockerfile":
		return "dockerfile"

	case "build", "workspace":
		return "starlark"
	}

	return ""
}

// getLanguageFromExtension returns the LSP mode for the
// provided file extension (e.g. "jsx")
//
// Cherry picked from https://github.com/isagalaev/highlight.js/tree/master/src/languages
// and https://github.com/github/linguist/blob/master/lib/linguist/languages.yml.
func getLanguageFromExtension(extension string) string {
	switch strings.ToLower(extension) {
	// Ada
	case "adb", "ada", "ads":
		return "ada"

	// Apex
	case "cls", "apex", "trigger":
		return "apex"

	// Actionscript
	case "as":
		return "actionscript"

	// Apache
	case "apacheconf":
		return "apache"

		// Applescript
	case "applescript", "scpt":
		return "applescript"

		// Bash
	case "sh", "bash", "zsh":
		return "shell"

		// Clojure
	case "clj", "cljs", "cljx":
		return "clojure"

		// CSS
	case "css":
		return "css"

		// CMake
	case "cmake", "cmake.in", "in": // TODO(john): hack b/c we don"t properly parse extensions w/ "." in them
		return "cmake"

		// Coffeescript
	case "coffee", "cake", "cson", "cjsx", "iced":
		return "coffescript"

		// C#
	case "cs", "csx":
		return "csharp"

		// C/C++
	case "c", "cc", "cpp", "cxx", "c++", "h++", "hh", "h", "hpp", "pc", "pcc":
		return "cpp"

		// CUDA
	case "cu", "cuh":
		return "cuda"

		// Dart
	case "dart":
		return "dart"

		// Diff
	case "diff", "patch":
		return "diff"

		// Django
	case "jinja":
		return "django"

		// DOS
	case "bat", "cmd":
		return "dos"

		// Elixir
	case "ex", "exs":
		return "elixir"

		// Elm
	case "elm":
		return "elm"

		// Erlang
	case "erl":
		return "erlang"

		// Fortran
	case "f", "for", "frt", "fr", "forth", "4th", "fth":
		return "fortran"

		// F#
	case "fs":
		return "fsharp"

		// Go
	case "go":
		return "go"

		// GraphQL
	case "graphql":
		return "graphql"

		// Groovy
	case "groovy":
		return "groovy"

		// HAML
	case "haml":
		return "haml"

		// Handlebars
	case "hbs", "handlebars":
		return "handlebars"

		// Haskell
	case "hs", "hsc":
		return "haskell"

		// HTML
	case "htm", "html", "xhtml":
		return "html"

		// INI
	case "ini", "cfg", "prefs", "pro", "properties":
		return "ini"

		// Java
	case "java":
		return "java"

		// JavaScript
	case "js", "jsx", "es", "es6", "mjs", "jss", "jsm":
		return "javascript"

		// JSON
	case "json", "sublime_metrics", "sublime_session", "sublime-keymap", "sublime-mousemap", "sublime-project", "sublime-settings", "sublime-workspace":
		return "json"

		// Jsonnet
	case "jsonnet", "libsonnet":
		return "jsonnet"

		// Julia
	case "jl":
		return "julia"

		// Kotlin
	case "kt", "ktm", "kts":
		return "kotlin"

		// Less
	case "less":
		return "less"

		// Lisp
	case "lisp", "asd", "cl", "lsp", "l", "ny", "podsl", "sexp", "el":
		return "lisp"

		// Lua
	case "lua", "fcgi", "nse", "pd_lua", "rbxs", "wlua":
		return "lua"

		// Makefile
	case "mk", "mak":
		return "makefile"

		// Markdown
	case "md", "mkdown", "mkd":
		return "markdown"

		// nginx
	case "nginxconf":
		return "nginx"

		// Objective-C
	case "m", "mm":
		return "objectivec"

		// OCaml
	case "ml", "eliom", "eliomi", "ml4", "mli", "mll", "mly", "re": // reason has the same language server as ocaml
		return "ocaml"

		// Pascal
	case "p", "pas", "pp":
		return "pascal"

		// Perl
	case "pl", "al", "cgi", "perl", "ph", "plx", "pm", "pod", "psgi", "t":
		return "perl"

		// PHP
	case "php", "phtml", "php3", "php4", "php5", "php6", "php7", "phps":
		return "php"

		// Powershell
	case "ps1", "psd1", "psm1":
		return "powershell"

		// Proto
	case "proto":
		return "protobuf"

		// Python
	case "py", "pyc", "pyd", "pyo", "pyw", "pyz":
		return "python"

		// R
	case "r", "rd", "rsx":
		return "r"
	case "repro":
		return "reprolang"

		// Ruby
	case "rb",
		"builder",
		"eye",
		"gemspec",
		"god",
		"jbuilder",
		"mspec",
		"pluginspec",
		"podspec",
		"rabl",
		"rake",
		"rbuild",
		"rbw",
		"rbx",
		"ru",
		"ruby",
		"spec",
		"thor",
		"watchr":
		return "ruby"

		// Rust
	case "rs", "rs.in":
		return "rust"

		// SASS
	case "sass", "scss":
		return "scss"

		// Scala
	case "sbt", "sc", "scala":
		return "scala"

		// Starlark
	case "bzl", "bazel":
		return "starlark"

		// Strato
	case "strato":
		return "strato"

		// Scheme
	case "scm", "sch", "sls", "sps", "ss":
		return "scheme"

		// Smalltalk
	case "st":
		return "smalltalk"

		// SQL
	case "sql":
		return "sql"

		// Stylus
	case "styl":
		return "stylus"

		// Swift
	case "swift":
		return "swift"

		// Thrift
	case "thrift":
		return "thrift"

		// TypeScript
	case "ts", "tsx":
		return "typescript"

		// Twig
	case "twig":
		return "twig"

		// Visual Basic
	case "vb":
		return "vbnet"
	case "vbs":
		return "vbscript"

		// Verilog, including SystemVerilog
	case "v", "veo", "sv", "svh", "svi":
		return "verilog"

		// VHDL
	case "vhd", "vhdl":
		return "vhdl"

		// VIM
	case "vim":
		return "vim"

		// XLSG
	case "xlsg":
		return "xlsg"

		// XML
	case "xml",
		"adml",
		"admx",
		"ant",
		"axml",
		"builds",
		"ccxml",
		"clixml",
		"cproject",
		"csl",
		"csproj",
		"ct",
		"dita",
		"ditamap",
		"ditaval",
		"dll.config",
		"dotsettings",
		"filters",
		"fsproj",
		"fxml",
		"glade",
		"gml",
		"grxml",
		"iml",
		"ivy",
		"jelly",
		"jsproj",
		"kml",
		"launch",
		"mdpolicy",
		"mjml",
		"mod",
		"mxml",
		"nproj",
		"nuspec",
		"odd",
		"osm",
		"pkgproj",
		"plist",
		"props",
		"ps1xml",
		"psc1",
		"pt",
		"rdf",
		"resx",
		"rss",
		"scxml",
		"sfproj",
		"srdf",
		"storyboard",
		"stTheme",
		"sublime-snippet",
		"targets",
		"tmCommand",
		"tml",
		"tmLanguage",
		"tmPreferences",
		"tmSnippet",
		"tmTheme",
		"ui",
		"urdf",
		"ux",
		"vbproj",
		"vcxproj",
		"vsixmanifest",
		"vssettings",
		"vstemplate",
		"vxml",
		"wixproj",
		"wsdl",
		"wsf",
		"wxi",
		"wxl",
		"wxs",
		"x3d",
		"xacro",
		"xaml",
		"xib",
		"xlf",
		"xliff",
		"xmi",
		"xml.list",
		"xproj",
		"xsd",
		"xspec",
		"xul",
		"zcml":
		return "xml"

	case "zig":
		return "zig"

		// YAML
	case "yml", "yaml":
		return "yaml"
	}
	return ""
}

func getPathExtension(path string) string {
	pathSplit := strings.Split(path, ".")
	if len(pathSplit) == 1 {
		return ""
	}
	if len(pathSplit) == 2 && pathSplit[0] == "" {
		return "" // e.g. .gitignore
	}
	return strings.ToLower(pathSplit[len(pathSplit)-1])
}
