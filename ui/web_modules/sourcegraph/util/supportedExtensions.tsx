import { getModes } from "sourcegraph/util/features";

const modes = getModes();

const ignoredExtensions = new Set<string>(["md"]);
const supportedExtensions = new Set<string>(["go", "js", "jsx", "ts", "tsx"]);
if (modes.has("css")) {
	supportedExtensions.add("css");
	supportedExtensions.add("less");
	supportedExtensions.add("scss");
}
if (modes.has("php")) {
	supportedExtensions.add("php");
}
if (modes.has("python")) {
	supportedExtensions.add("py");
}
if (modes.has("java")) {
	supportedExtensions.add("java");
}

export function isSupportedExtension(ext: string): boolean {
	return supportedExtensions.has(ext);
}

// ignored extensions, like md, will not trigger a warning banner
export function isIgnoredExtension(ext: string): boolean {
	return ignoredExtensions.has(ext);
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
