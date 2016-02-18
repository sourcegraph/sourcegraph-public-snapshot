let context;

// TODO(autotest) support window object.
if (typeof window !== "undefined") {
	context = {
		currentUser: window._currentUser,
		csrfToken: window._csrfToken,
		isMothership: window._isMothership,
	};
} else {
	context = {
		currentUser: null,
		csrfToken: "",
		isMothership: false,
	};
}

export default context;
