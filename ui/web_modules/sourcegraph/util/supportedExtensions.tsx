export function typescriptSupported(): boolean {
	if (typeof global.window === "undefined") {
		return false;
	}
	return Boolean(window && window.localStorage["xlangTypescript"]);
}

let _exts = ["go"];
if (typescriptSupported()) {
	_exts.push("ts");
	_exts.push("tsx");
}

export const supportedExtensions = _exts;

export function isSupportedExtension(ext: string): boolean {
	return supportedExtensions.indexOf(ext) !== -1;
}

export function isSupportedMode(modeId: string): boolean {
	if (modeId === "go") {
		return true;
	}
	return typescriptSupported() && modeId === "typescript";
}

export function getPathExtension(path: string): string | null {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1 ) { return null; };
	if (pathSplit.length === 2 && pathSplit[0] === "") { return null; }; // e.g. .gitignore
	return pathSplit[pathSplit.length - 1].toLowerCase();
}
