// Override the electron import path. Workbench technically requires electron,
// but we don't use any functionality from it.

export const remote = {
	getCurrentWindow: () => {
		return 1;
	},
};
