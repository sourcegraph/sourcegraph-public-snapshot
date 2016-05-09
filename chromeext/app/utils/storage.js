function saveState(state) {
	delete state.text; // migrate unused data
	chrome.runtime.sendMessage(null, {type: "set", state: JSON.stringify(state)}, {});
}

export default function() {
	return (next) => (reducer, initialState) => {
		const store = next(reducer, initialState);
		store.subscribe(() => {
			const state = store.getState();
			saveState(state);
			// You may include other side effects like `chrome.browserAction.setBadgeText`,
			// event logging, etc.
		});
		return store;
	};
}
