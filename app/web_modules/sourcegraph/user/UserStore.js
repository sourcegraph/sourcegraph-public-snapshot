import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as UserActions from "sourcegraph/user/UserActions";

export class UserStore extends Store {
	reset(data?: any) {
		this.activeAccessToken = data && data.activeAccessToken ? data.activeAccessToken : null;
		this.activeGitHubToken = data && data.activeGitHubToken ? data.activeGitHubToken : null;
		this.authInfo = deepFreeze({
			byAccessToken: data && data.authInfo ? data.authInfo.byAccessToken : {},
			get(accessToken) {
				return this.byAccessToken[accessToken] || null;
			},
		});
		this.users = deepFreeze({
			byUID: data && data.users ? data.users.byUID : {},
			get(uid) {
				return this.byUID[uid] || null;
			},
		});
		this.emails = deepFreeze({
			byUID: data && data.emails ? data.emails.byUID : {},
			get(uid) {
				return this.byUID[uid] || null;
			},
		});
		this.pendingAuthActions = deepFreeze({
			content: data && data.pendingAuthActions ? data.pendingAuthActions.content : {},
			get(state) {
				return this.content[state] || null;
			},
		});
		this.authResponses = deepFreeze({
			content: data && data.authResponses ? data.authResponses.content : {},
			get(state) {
				return this.content[state] || null;
			},
		});
		this.settings = deepFreeze({
			content: data && data.settings ? data.settings.content : {},
			get(/* no args, stored on user's browser */) {
				return this.content || null;
			},
		});
	}

	toJSON() {
		return {
			activeAccessToken: this.activeAccessToken,
			activeGitHubToken: this.activeGitHubToken,
			authInfo: this.authInfo,
			users: this.users,
			emails: this.emails,
			pendingAuthActions: this.pendingAuthActions,
			authResponses: this.authResponses,
			settings: this.settings,
		};
	}

	// activeAuthInfo returns the AuthInfo object for the active user, if there
	// is one. Otherwise it returns null.
	activeAuthInfo() {
		if (!this.activeAccessToken) return null;
		return this.authInfo.get(this.activeAccessToken);
	}

	// activeUser returns the User object for the active user, if there is one
	// and if the User object is already persisted in the store. Otherwise it
	// returns null.
	activeUser() {
		const authInfo = this.activeAuthInfo();
		if (!authInfo || !authInfo.UID) return null;
		const user = this.users.get(authInfo.UID);
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
		// Using instanceof checks instead of switching on action.constructor
		// lets Flow understand the type constraints, so we should move the
		// rest of the switch-case bodies to this scheme.

		if (action instanceof UserActions.FetchedAuthInfo) {
			this.authInfo = deepFreeze({
				...this.authInfo,
				byAccessToken: {
					...this.authInfo.byAccessToken,
					[action.accessToken]: action.authInfo,
				},
			});
			this.__emitChange();
			return;
		} else if (action instanceof UserActions.FetchedUser) {
			this.users = deepFreeze({
				...this.users,
				byUID: {
					...this.users.byUID,
					[action.uid]: action.user,
				},
			});
			this.__emitChange();
			return;
		} else if (action instanceof UserActions.FetchedEmails) {
			this.emails = deepFreeze({
				...this.emails,
				byUID: {
					...this.emails.byUID,
					[action.uid]: action.emails,
				},
			});
			this.__emitChange();
			return;
		} else if (action instanceof UserActions.FetchedGitHubToken) {
			this.activeGitHubToken = action.token;
			this.__emitChange();
			return;
		} else if (action instanceof UserActions.UpdateSettings) {
			this.settings = deepFreeze({...this.settings, content: action.settings});
			this.__emitChange();
			return;
		}


		switch (action.constructor) {
		case UserActions.SubmitSignup: {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					signup: true,
				},
			});
			break;
		}
		case UserActions.SubmitLogin: {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					login: true,
				},
			});
			break;
		}
		case UserActions.SubmitLogout: {
			this._resetAuth();
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					logout: true,
				},
			});
			break;
		}
		case UserActions.SubmitForgotPassword: {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					forgot: true,
				},
			});
			break;
		}
		case UserActions.SubmitResetPassword: {
			this.pendingAuthActions = deepFreeze({
				...this.pendingAuthActions,
				content: {
					...this.pendingAuthActions.content,
					reset: true,
				},
			});
			break;
		}
		case UserActions.SignupCompleted: {
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
			break;
		}
		case UserActions.LoginCompleted: {
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
			break;
		}
		case UserActions.LogoutCompleted: {
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
			break;
		}
		case UserActions.ForgotPasswordCompleted: {
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
			break;
		}
		case UserActions.ResetPasswordCompleted: {
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
			break;
		}
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new UserStore(Dispatcher.Stores);
