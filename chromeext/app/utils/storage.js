function saveState(state) {
	delete state.text; // migrate unused data
	chrome.storage.local.set({state: JSON.stringify(state)});
}

export default function() {
	return (next) => (reducer, initialState) => {
		const store = next(reducer, initialState);
		store.subscribe(() => {
			const state = store.getState();
			console.log(state);
			saveState(state);
			// You may include other side effects like `chrome.browserAction.setBadgeText`,
			// event logging, etc.
		});
		return store;
	};
}
