// @flow

import UserStore from "sourcegraph/user/UserStore";
import {forEach as forEachStore, reset as resetStores} from "sourcegraph/init/stores";

// This module, as a side effect of being imported, clears the authentication
// information and resets all stores to their initial state (with no data) when
// the authenticated user changes (after login, signup, or logout).

let lastActiveAccessToken = UserStore.activeAccessToken;

UserStore.addListener(() => {
	if (UserStore.activeAccessToken !== lastActiveAccessToken) {
		lastActiveAccessToken = UserStore.activeAccessToken;

		// Keep some UserStore data related to the status of the just-occurred
		// auth change and the new/empty access token.
		const {activeAccessToken, activeGitHubToken, pendingAuthActions, authResponses} = UserStore.toJSON();
		resetStores({
			UserStore: {activeAccessToken, activeGitHubToken, pendingAuthActions, authResponses},
		});

		forEachStore((store) => {
			store.__emitChange();
		});
	}
});
