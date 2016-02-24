let context;

// TODO(autotest) support window object.
if (typeof window !== "undefined") {
	context = {
		currentUser: window._currentUser,
		csrfToken: window._csrfToken,
		isMothership: window._isMothership,
		cacheControl: window._cacheControl || null,
	};
} else {
	context = {
		currentUser: null,
		csrfToken: "",
		isMothership: false,
		cacheControl: null,
	};
}

export default context;
