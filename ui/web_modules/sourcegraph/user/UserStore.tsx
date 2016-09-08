import {AuthInfo, User} from "sourcegraph/api";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Store} from "sourcegraph/Store";
import {Settings} from "sourcegraph/user";
import * as UserActions from "sourcegraph/user/UserActions";
import {deepFreeze, mergeAndDeepFreeze} from "sourcegraph/util/deepFreeze";

class UserStoreClass extends Store<any> {
	activeAccessToken: string | null;
	authInfos: {[key: string]: AuthInfo | null};
	users: {[key: string]: User};
	pendingAuthActions: {[key: string]: boolean};
	authResponses: {[key: string]: any};
	settings: Settings;

	reset(): void {
		this.activeAccessToken = null;
		this.authInfos = deepFreeze({});
		this.users = deepFreeze({});
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

	__onDispatch(action: UserActions.Action): void {
		if (action instanceof UserActions.FetchedAuthInfo) {
			this.authInfos = mergeAndDeepFreeze(this.authInfos, {[action.accessToken]: action.authInfo});

		} else if (action instanceof UserActions.FetchedUser) {
			this.users = mergeAndDeepFreeze(this.users, {[action.uid]: action.user});

		} else if (action instanceof UserActions.UpdateSettings) {
			if (global.window) { window.localStorage.setItem("userSettings", JSON.stringify(action.settings)); }
			this.settings = deepFreeze(Object.assign({}, this.settings, action.settings));

		} else if (action instanceof UserActions.SubmitSignup) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {signup: true});

		} else if (action instanceof UserActions.SubmitLogin) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {login: true});

		} else if (action instanceof UserActions.SubmitForgotPassword) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {forgot: true});

		} else if (action instanceof UserActions.SubmitResetPassword) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {reset: true});

		} else if (action instanceof UserActions.SignupCompleted) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {signup: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {signup: action.resp});

		} else if (action instanceof UserActions.LoginCompleted) {
			this.pendingAuthActions = mergeAndDeepFreeze(this.pendingAuthActions, {login: false});
			this.authResponses = mergeAndDeepFreeze(this.authResponses, {login: action.resp});

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

export const UserStore = new UserStoreClass(Dispatcher.Stores);
