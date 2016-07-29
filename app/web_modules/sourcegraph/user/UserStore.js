import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as UserActions from "sourcegraph/user/UserActions";

export class UserStore extends Store {
	reset() {
		this.activeAccessToken = null;
		this.activeGitHubToken = null;
		this.authInfos = deepFreeze({});
		this.users = deepFreeze({});
		this.emails = deepFreeze({});
		this.pendingAuthActions = deepFreeze({
			content: {},
			get(state) {
				return this.content[state] || null;
			},
		});
		this.authResponses = deepFreeze({
			content: {},
			get(state) {
				return this.content[state] || null;
			},
		});

		if (global.window) {
			let storedUserSettings = window.localStorage.getItem("userSettings");
			if (!storedUserSettings) {
				storedUserSettings = {
					search: {
						languages: ["golang"],
						scope: {
							popular: true,
							public: false,
							private: false,
							repo: true,
						},
					},
				};
			} else {
				storedUserSettings = JSON.parse(storedUserSettings);
			}

			this.settings = deepFreeze({
				content: storedUserSettings,
				get(/* no args, stored on user's browser */) {
					return this.content || null;
				},
			});
		}
	}

	// activeAuthInfo returns the AuthInfo object for the active user, if there
	// is one. Otherwise it returns null.
	activeAuthInfo() {
		if (!this.activeAccessToken) return null;
		return this.authInfos[this.activeAccessToken] || null;
	}

	// activeUser returns the User object for the active user, if there is one
	// and if the User object is already persisted in the store. Otherwise it
	// returns null.
	activeUser() {
		const authInfo = this.activeAuthInfo();
		if (!authInfo || !authInfo.UID) return null;
		const user = this.users[authInfo.UID];
		return user && !user.Error ? user : null;
	}

	// _resetAuth causes resetOnAuthChange's listener to be called, which clears
	// all store data after an auth change (login/signup/logout). This is so that
	// users don't see data that was fetched with the auth of the previous user signed
	// into the app in their browser.
	_resetAuth() {
		this.activeAccessToken = null;
	}

	__onDispatch(action) {
		if (action instanceof UserActions.FetchedAuthInfo) {
			this.authInfos = deepFreeze(Object.assign({}, this.authInfos, {[action.accessToken]: action.authInfo}));

		} else if (action instanceof UserActions.FetchedUser) {
			this.users = deepFreeze(Object.assign({}, this.users, {[action.uid]: action.user}));

		} else if (action instanceof UserActions.FetchedEmails) {
			this.emails = deepFreeze(Object.assign({}, this.emails, {[action.uid]: action.emails}));

		} else if (action instanceof UserActions.FetchedGitHubToken) {
			this.activeGitHubToken = action.token;

		} else if (action instanceof UserActions.UpdateSettings) {
			if (global.window) window.localStorage.setItem("userSettings", JSON.stringify(action.settings));
			this.settings = deepFreeze({...this.settings, content: action.settings});

		} else if (action instanceof UserActions.SubmitSignup) {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					signup: true,
				},
			});

		} else if (action instanceof UserActions.SubmitLogin) {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					login: true,
				},
			});

		} else if (action instanceof UserActions.SubmitLogout) {
			this._resetAuth();
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					logout: true,
				},
			});

		} else if (action instanceof UserActions.SubmitForgotPassword) {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					forgot: true,
				},
			});

		} else if (action instanceof UserActions.SubmitResetPassword) {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					reset: true,
				},
			});

		} else if (action instanceof UserActions.SignupCompleted) {
			this._resetAuth();
			if (action.resp && action.resp.Success) this.activeAccessToken = action.resp.AccessToken;
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					signup: false,
				},
			});
			this.authResponses = deepFreeze({
				...this.authResponses,
				content: {
					...this.authResponses.content,
					signup: action.resp,
				},
			});

		} else if (action instanceof UserActions.LoginCompleted) {
			this._resetAuth();
			if (action.resp && action.resp.Success) this.activeAccessToken = action.resp.AccessToken;
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					login: false,
				},
			});
			this.authResponses = deepFreeze({
				...this.authResponses,
				content: {
					...this.authResponses.content,
					login: action.resp,
				},
			});

		} else if (action instanceof UserActions.LogoutCompleted) {
			this._resetAuth();
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					logout: false,
				},
			});
			this.authResponses = deepFreeze({
				...this.authResponses,
				content: {
					...this.authResponses.content,
					logout: action.resp,
				},
			});

		} else if (action instanceof UserActions.ForgotPasswordCompleted) {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					forgot: false,
				},
			});
			this.authResponses = deepFreeze({
				...this.authResponses,
				content: {
					...this.authResponses.content,
					forgot: action.resp,
				},
			});

		} else if (action instanceof UserActions.ResetPasswordCompleted) {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					reset: false,
				},
			});
			this.authResponses = deepFreeze({
				...this.authResponses,
				content: {
					...this.authResponses.content,
					reset: action.resp,
				},
			});

		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new UserStore(Dispatcher.Stores);
