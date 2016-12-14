// Workbench requires native-keymap module, but we override it after startup
// with a standalone version, so provide a stub here.
export function getKeyMap(): any {
	return {};
}
