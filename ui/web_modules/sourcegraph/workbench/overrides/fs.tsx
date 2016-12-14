// This overrides the node fs module. It is used by the Workbench to find
// configuration. All normal file access goes through the FileService. VSCode
// uses the default configuration if it can't find this one, so returning
// nothing is acceptable here.

export function exists(): boolean {
	return false;
}

export function lstat(): any {
	return {};
}

export function readFile(path: string): string {
	return "";
}
