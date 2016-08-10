// tslint:disable: typedef ordered-imports

import {UserStore} from "sourcegraph/user/UserStore";
import {forEach as forEachStore, reset as resetStores} from "sourcegraph/init/stores";

// resetOnAuthChange clears the authentication information in UserStore
// and resets all stores to their initial state (with no data) when
// the authenticated user changes (after login, signup, or logout).
export function resetOnAuthChange() {
	let lastActiveAccessToken = UserStore.activeAccessToken;
	if (typeof lastActiveAccessToken === "undefined") {
		throw new Error("resetOnAuthChange must be called after the initial UserStore.activeAccessToken has been set (to null or to the access token string). Usually this means it must be called after setting the JS context.");
	}

	UserStore.addListener(() => {
		if (UserStore.activeAccessToken !== lastActiveAccessToken) {
			lastActiveAccessToken = UserStore.activeAccessToken;

			// Keep some UserStore data related to the status of the just-occurred
			// auth change and the new/empty access token.
			const activeAccessToken = UserStore.activeAccessToken;
			const activeGitHubToken = UserStore.activeGitHubToken;
			const pendingAuthActions = UserStore.pendingAuthActions;
			const authResponses = UserStore.authResponses;
			resetStores();
			UserStore.activeAccessToken = activeAccessToken;
			UserStore.activeGitHubToken = activeGitHubToken;
			UserStore.pendingAuthActions = pendingAuthActions;
			UserStore.authResponses = authResponses;

			forEachStore((store) => {
				store.__emitChange();
			});
		}
	});
}
