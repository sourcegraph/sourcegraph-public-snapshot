import * as Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import {UserStore} from "sourcegraph/user/UserStore";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

class UserBackendClass {
	fetch: any;

	constructor() {
		this.fetch = defaultFetch;
	}

	__onDispatch(payload: UserActions.Action): void {
		if (payload instanceof UserActions.WantAuthInfo) {
			const action = payload;
			if (!UserStore.authInfos[action.accessToken]) {
				this.fetch("/.api/auth-info")
					.then(checkStatus)
					.then((resp) => resp.json())
					.then(function(data: any): void {
						// The user and emails might've been optimistically included in the API response.
						let user = data.IncludedUser;
						if (user) { delete data.IncludedUser; }
						let emails = data.IncludedEmails;
						if (emails) { delete data.IncludedEmails; }
						let token = data.GitHubToken;
						if (token) { delete data.GitHubToken; }

						// Dispatch FetchedUser before FetchedAuthInfo because it's common for components
						// to dispatch a WantUser when the auth info is received, and dispatching FetchedUser
						// first means that WantUser will immediately return because the data is already there.
						if (user && data.UID) {
							Dispatcher.Stores.dispatch(new UserActions.FetchedUser(data.UID, user));
						}

						Dispatcher.Stores.dispatch(new UserActions.FetchedAuthInfo(action.accessToken, data));

						if (emails && data.UID) {
							Dispatcher.Stores.dispatch(new UserActions.FetchedEmails(data.UID, emails));
						}
						if (token && data.UID) {
							Dispatcher.Stores.dispatch(new UserActions.FetchedGitHubToken(data.UID, token));
						}
					}, function(err: any): void { console.error(err); });
			}
		}

		if (payload instanceof UserActions.WantUser) {
			const action = payload;
			if (!UserStore.users[action.uid]) {
				this.fetch(`/.api/users/${action.uid}$`) // trailing "$" indicates UID lookup (not login/username)
					.then(checkStatus)
					.then((resp) => resp.json())
					.then(function(data: any): void {
						Dispatcher.Stores.dispatch(new UserActions.FetchedUser(action.uid, data));
					}, function(err: any): void { console.error(err); });
			}
		}

		if (payload instanceof UserActions.WantEmails) {
			const action = payload;
			if (!UserStore.emails[action.uid]) {
				this.fetch(`/.api/users/${action.uid}$/emails`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.then(function(data: any): void {
						Dispatcher.Stores.dispatch(new UserActions.FetchedEmails(action.uid, data && data.EmailAddrs ? data.EmailAddrs : []));
					}, function(err: any): void { console.error(err); });
			}
		}

		if (payload instanceof UserActions.SubmitSignup) {
			const action = payload;
			this.fetch(`/.api/join`, {
				method: "POST",
				body: JSON.stringify({
					Login: action.login,
					Password: action.password,
					Email: action.email,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.SignupCompleted(action.email, data));
				});
		}

		if (payload instanceof UserActions.SubmitLogin) {
			const action = payload;
			this.fetch(`/.api/login`, {
				method: "POST",
				body: JSON.stringify({
					Login: action.login,
					Password: action.password,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.LoginCompleted(data));
				});
		}

		if (payload instanceof UserActions.SubmitLogout) {
			this.fetch(`/.api/logout`, {
				method: "POST",
				body: JSON.stringify({}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.LogoutCompleted(data));
					// Redirect on logout.
					if (data.Success) {
						window.location.href = "/#loggedout";
					}
				});
		}

		if (payload instanceof UserActions.SubmitForgotPassword) {
			const action = payload;
			this.fetch(`/.api/forgot`, {
				method: "POST",
				body: JSON.stringify({
					Email: action.email,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.ForgotPasswordCompleted(data));
				});
		}

		if (payload instanceof UserActions.SubmitResetPassword) {
			const action = payload;
			this.fetch(`/.api/reset`, {
				method: "POST",
				body: JSON.stringify({
					Password: action.password,
					ConfirmPassword: action.confirmPassword,
					Token: action.token,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.ResetPasswordCompleted(data));
				});
		}

		if (payload instanceof UserActions.SubmitBetaSubscription) {
			const action = payload;
			this.fetch(`/.api/beta-subscription`, {
				method: "POST",
				body: JSON.stringify({
					Email: action.email,
					FirstName: action.firstName,
					LastName: action.lastName,
					Languages: action.languages,
					Editors: action.editors,
					Message: action.message,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.BetaSubscriptionCompleted(data));
				});
		}
	}
};

export const UserBackend = new UserBackendClass();
Dispatcher.Backends.register(UserBackend.__onDispatch.bind(UserBackend));
