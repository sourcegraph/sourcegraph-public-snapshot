function saveAccessToken(state: {accessToken: string | null}): void {
	chrome.runtime.sendMessage({type: "set", state: JSON.stringify(state)}, {});
}

export function saveAccessTokenMiddleware(next: any): any {
	return (reducer, initialState) => {
		const store = next(reducer, initialState);
		store.subscribe(() => {
			const state = store.getState();
			saveAccessToken({accessToken: state.accessToken});
		});
		return store;
	};
}
