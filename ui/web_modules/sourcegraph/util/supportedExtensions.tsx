import { modes } from "sourcegraph/editor/modes";

const supportedExtensions = new Set<string>(["go", "js", "jsx", "ts", "tsx", "c", "h", "py"]);
if (modes.has("php")) {
	supportedExtensions.add("php");
}

export function isSupportedExtension(ext: string): boolean {
	return supportedExtensions.has(ext);
}

export function isSupportedMode(modeId: string): boolean {
	return modes.has(modeId);
}

export function getPathExtension(path: string): string | null {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1) { return null; };
	if (pathSplit.length === 2 && pathSplit[0] === "") { return null; }; // e.g. .gitignore
	return pathSplit[pathSplit.length - 1].toLowerCase();
}
