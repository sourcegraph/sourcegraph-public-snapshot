import { SourcegraphURL } from "util/types";

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
		default:
			return "unknown";
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

export function parseURL(loc: Location = window.location): SourcegraphURL {

	const urlsplit = loc.pathname.slice(1).split("/");
	if (urlsplit.length < 3 && urlsplit[0] !== "github.com") {
		return {};
	}

	let uri = urlsplit.slice(0, 3).join("/");
	let rev: string | undefined;
	let path: string | undefined;
	const uriSplit = uri.split("@");
	if (uriSplit.length > 0) {
		uri = uriSplit[0];
		rev = uriSplit[1];
	}

	if (loc.pathname.indexOf("/-/blob/") !== -1) {
		path = urlsplit.slice(5).join("/");
	}

	return { uri, rev, path };
}

export function urlToBlob(loc: { uri: string, rev?: string, path: string, line?: number, char?: number, refs?: "all" | "local" | "external" }): string {
	let url = `/${loc.uri}${loc.rev ? "@" + loc.rev : ""}/-/blob/${loc.path}`;
	if (loc.line) { // construct hash w/ format #L[line][:char][$references[:local|external]]
		url += "#L" + loc.line;
		if (loc.char) {
			url += ":" + loc.char;
			if (loc.refs) {
				url += "$references";
				if (loc.refs === "local") {
					url += ":local";
				}
				if (loc.refs === "external") {
					url += ":external";
				}
			}
		}
	}
	return url;
}
