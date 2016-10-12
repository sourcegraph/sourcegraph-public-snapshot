export const supportedExtensions = ["go", "ts", "tsx"];

export function isSupportedExtension(ext: string): boolean {
	return supportedExtensions.indexOf(ext) !== -1;
}

export function isSupportedMode(modeId: string): boolean {
	return modeId === "go" || modeId === "typescript";
}

export function getPathExtension(path: string): string | null {
	const pathSplit = path.split(".");
	if (pathSplit.length === 1 ) { return null; };
	if (pathSplit.length === 2 && pathSplit[0] === "") { return null; }; // e.g. .gitignore
	return pathSplit[pathSplit.length - 1].toLowerCase();
}
