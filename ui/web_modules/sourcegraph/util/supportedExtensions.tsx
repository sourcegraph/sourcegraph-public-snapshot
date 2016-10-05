// Adapted from client/browser-ext/app/utils/index.js

export const supportedExtensions = [
	"go", "ts", "tsx", //"java", "js", "jsx"
];

export function getPathExtension(path: string): string | null {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1 ) { return null; };
	if (pathSplit.length === 2 && pathSplit[0] === "") { return null; }; // e.g. .gitignore
	return pathSplit[pathSplit.length - 1].toLowerCase();
}
