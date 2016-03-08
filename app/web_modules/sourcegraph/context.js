let context;

// TODO(autotest) support window object.
if (typeof window !== "undefined") {
	context = {
		currentUser: window._currentUser,
		csrfToken: window._csrfToken,
		cacheControl: window._cacheControl || null,
	};
} else {
	context = {
		currentUser: null,
		csrfToken: "",
		cacheControl: null,
	};
}

export default context;
