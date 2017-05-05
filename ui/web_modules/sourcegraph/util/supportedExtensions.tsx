import { getModes } from "sourcegraph/util/features";

const modes = getModes();

const ignoredExtensions = new Set<string>(["md", "txt", "json", "yml"]);
const supportedExtensions = new Set<string>(["go", "java", "js", "jsx", "ts", "tsx"]);
const betaExtensions = new Set<string>([]);
if (modes.has("css")) {
	betaExtensions.add("css");
	betaExtensions.add("less");
	betaExtensions.add("scss");
}
if (modes.has("php")) {
	betaExtensions.add("php");
}
if (modes.has("python")) {
	betaExtensions.add("py");
}
if (modes.has("swift")) {
	betaExtensions.add("swift");
}

export function isSupportedExtension(ext: string): boolean {
	return supportedExtensions.has(ext);
}

// ignored extensions, like md, will not trigger a warning banner
export function isIgnoredExtension(ext: string): boolean {
	return ignoredExtensions.has(ext);
}

export function isBetaExtension(ext: string): boolean {
	return betaExtensions.has(ext);
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
