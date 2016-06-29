function saveState(state) {
	chrome.runtime.sendMessage(null, {type: "set", state: JSON.stringify(state)}, {});
}

export default function() {
	return (next) => (reducer, initialState) => {
		const store = next(reducer, initialState);
		store.subscribe(() => {
			const state = store.getState();
			saveState({accessToken: state.accessToken});
		});
		return store;
	};
}
