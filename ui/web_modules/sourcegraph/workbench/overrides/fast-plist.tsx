// Workbench imports the fast-plist module but Sourcegraph doesn't
// actually use it.
export function parse(): any {
	throw new Error("fast-plist: not implemented");
}
