import Dispatcher from "sourcegraph/Dispatcher";
import Store from "sourcegraph/Store";
import * as UserActions from "sourcegraph/user/UserActions";
import {AuthInfo, EmailAddr, ExternalToken, Settings, User} from "sourcegraph/user/index";
import {deepFreeze, mergeAndDeepFreeze} from "sourcegraph/util/deepFreeze";

export class UserStore extends Store<any> {
	activeAccessToken: string | null;
	activeGitHubToken: ExternalToken | null;
	authInfos: {[key: string]: AuthInfo | null};
	users: {[key: string]: User};
	emails: {[key: string]: EmailAddr[]};
	pendingAuthActions: {[key: string]: boolean};
	authResponses: {[key: string]: any};
	settings: Settings;

	reset(): void {
		this.activeAccessToken = null;
		this.activeGitHubToken = null;
		this.authInfos = deepFreeze({});
		this.users = deepFreeze({});
		this.emails = deepFreeze({});
		this.pendingAuthActions = deepFreeze({});
		this.authResponses = deepFreeze({});

		if (global.window) {
			let storedUserSettings = window.localStorage.getItem("userSettings");
			if (storedUserSettings) {
				this.settings = deepFreeze(JSON.parse(storedUserSettings) as Settings);
			} else {
				this.settings = deepFreeze({
					search: {
						languages: ["golang"],
						scope: {
							popular: true,
							public: false,
							private: false,
							repo: true,
						},
					},
				});
			}
		}
	}

	// activeAuthInfo returns the AuthInfo object for the active user, if there
	// is one. Otherwise it returns null.
	activeAuthInfo(): AuthInfo | null {
		if (!this.activeAccessToken) { return null; }
		return this.authInfos[this.activeAccessToken] || null;
	}

	// activeUser returns the User object for the active user, if there is one
	// and if the User object is already persisted in the store. Otherwise it
	// returns null.
	activeUser(): User | null {
		const authInfo = this.activeAuthInfo();
		if (!authInfo || !authInfo.UID) { return null; }
		return this.users[authInfo.UID] || null;
	}

	// _resetAuth causes resetOnAuthChange's listener to be called, which clears
	// all store data after an auth change (login/signup/logout). This is so that
	// users don't see data that was fetched with the auth of the previous user signed
	// into the app in their browser.
	_resetAuth(): void {
		this.activeAccessToken = null;
	}

	__onDispatch(action: UserActions.Action): void {
		if (action instanceof UserActions.FetchedAuthInfo) {
			this.authInfos = mergeAndDeepFreeze(this.authInfos, {[action.accessToken]: action.authInfo});

		} else if (action instanceof UserActions.FetchedUser) {
			this.users = mergeAndDeepFreeze(this.users, {[action.uid]: action.user});

		} else if (action instanceof UserActions.FetchedEmails) {
			this.emails = mergeAndDeepFreeze(this.emails, {[action.uid]: action.emails});

		} else if (action instanceof UserActions.FetchedGitHubToken) {
			this.activeGitHubToken = action.token;

		} else if (action instanceof UserActions.UpdateSettings) {
			if (global.window) { window.localStorage.setItem("userSettings", JSON.stringify(action.settings)); }
			this.settings = deepFreeze(Object.assign({}, this.settings, {content: action.settings}));

		} else if (action instanceof UserActions.SubmitSignup) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {signup: true});

		} else if (action instanceof UserActions.SubmitLogin) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {login: true});

		} else if (action instanceof UserActions.SubmitLogout) {
			this._resetAuth();
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {logout: true});

		} else if (action instanceof UserActions.SubmitForgotPassword) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {forgot: true});

		} else if (action instanceof UserActions.SubmitResetPassword) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {reset: true});

		} else if (action instanceof UserActions.SignupCompleted) {
			this._resetAuth();
			if (action.resp && action.resp.Success) { this.activeAccessToken = action.resp.AccessToken; }
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {signup: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {signup: action.resp});

		} else if (action instanceof UserActions.LoginCompleted) {
			this._resetAuth();
			if (action.resp && action.resp.Success) { this.activeAccessToken = action.resp.AccessToken; }
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {login: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {login: action.resp});

		} else if (action instanceof UserActions.LogoutCompleted) {
			this._resetAuth();
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {logout: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {logout: action.resp});

		} else if (action instanceof UserActions.ForgotPasswordCompleted) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {forgot: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {forgot: action.resp});

		} else if (action instanceof UserActions.ResetPasswordCompleted) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {reset: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {reset: action.resp});

		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new UserStore(Dispatcher.Stores);
