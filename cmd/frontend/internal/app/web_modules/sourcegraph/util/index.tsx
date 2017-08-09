/**
 * supportedExtensions are the file extensions
 * the extension will apply annotations to
 */
export const supportedExtensions = new Set<string>([
	"go", // Golang
	"ts", "tsx", // TypeScript
	"js", "jsx", // JavaScript
	"java", // Java
	"py", // Python
	"php", // PHP
]);

/**
 * getModeFromExtension returns the LSP mode for the
 * provided file extension (e.g. "jsx")
 */
export function getModeFromExtension(ext: string): string {
	switch (ext) {
		case "go":
			return "go";
		case "ts":
		case "tsx":
			return "typescript";
		case "js":
		case "jsx":
			return "javascript";
		case "java":
			return "java";
		case "py":
		case "pyc":
		case "pyd":
		case "pyo":
		case "pyw":
		case "pyz":
			return "python";
		case "php":
		case "phtml":
		case "php3":
		case "php4":
		case "php5":
		case "php6":
		case "php7":
		case "phps":
			return "php";
		case "htm":
		case "xhtml":
			return "html";
		case "erl":
			return "erlang";
		case "jsp":
			return "java";
		case "pl":
			return "perl";
		case "rss":
		case "atom":
		case "xsl":
		case "plist":
			return "xml";
		case "rb":
		case "builder":
		case "gemspec":
		case "podspec":
		case "thor":
			return "ruby";
		case "diff":
			return "patch";
		case "hs":
		case "icl":
			return "haskell";
		case "sh":
		case "zsh":
			return "bash";
		case "st":
			return "smalltalk";
		case "as":
			return "actionscript";
		case "apacheconf":
			return "apache";
		case "osacript":
			return "applescript";
		case "clj":
			return "clojure";
		case "cmake.in":
			return "cmake";
		case "coffee":
		case "cson":
		case "iced":
			return "coffescript";
		case "c++":
		case "h++":
		case "hh":
			return "cpp";
		case "jinja":
			return "django";
		case "bat":
		case "cmd":
			return "dos";
		case "fs":
			return "fsharp";
		case "hbs":
		case "html.hbs":
		case "html.handlebars":
			return "handlebars";
		case "json":
		case "sublime_metrics":
		case "sublime_session":
		case "sublime-keymap":
		case "sublime-mousemap":
		case "sublime-project":
		case "sublime-settings":
		case "sublime-workspace":
			return "json";
		case "mk":
			return "makefile";
		case "mak":
			return "makefile";
		case "md":
			return "markdown";
		case "mkdown":
			return "markdown";
		case "mkd":
			return "markdown";
		case "nginxconf":
			return "nginx";
		case "m":
		case "mm":
			return "objectivec";
		case "ml":
			return "ocaml";
		case "rs":
			return "rust";
		case "sci":
			return "scilab";
		case "vb":
			return "vbnet";
		case "vbs":
			return "vbscrip";
		default:
			return "";
	}
}

export function getPathExtension(path: string): string {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1) {
		return "";
	}
	if (pathSplit.length === 2 && pathSplit[0] === "") {
		return ""; // e.g. .gitignore
	}
	return pathSplit[pathSplit.length - 1].toLowerCase();
}
