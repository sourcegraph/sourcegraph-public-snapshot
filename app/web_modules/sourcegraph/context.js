let context;

// TODO(autotest) support window object.
if (typeof window !== "undefined") {
	context = {
		currentUser: window._currentUser,
		csrfToken: window._csrfToken,
	};
} else {
	context = {
		currentUser: null,
		csrfToken: "",
	};
}

export default context;
