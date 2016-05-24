import * as UserActions from "sourcegraph/user/UserActions";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";
import UserStore from "sourcegraph/user/UserStore";

const UserBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		// Using instanceof checks instead of switching on action.constructor
		// lets Flow understand the type constraints, so we should move the
		// rest of the switch-case bodies to this scheme.

		if (action instanceof UserActions.WantAuthInfo) {
			if (UserStore.authInfo.get(action.accessToken) === null) {
				trackPromise(
					UserBackend.fetch("/.api/auth-info")
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => {
							// The user and emails might've been optimistically included in the API response.
							let user = data.IncludedUser;
							if (user) delete data.IncludedUser;
							let emails = data.IncludedEmails;
							if (emails) delete data.IncludedEmails;
							let token = data.GitHubToken;
							if (token) delete data.GitHubToken;

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
						})
				);
			}
		} else if (action instanceof UserActions.WantUser) {
			if (UserStore.users.get(action.uid) === null) {
				trackPromise(
					UserBackend.fetch(`/.api/users/${action.uid}$`) // trailing "$" indicates UID lookup (not login/username)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new UserActions.FetchedUser(action.uid, data)))
				);
			}
		} else if (action instanceof UserActions.WantEmails) {
			if (UserStore.emails.get(action.uid) === null) {
				trackPromise(
					UserBackend.fetch(`/.api/users/${action.uid}$/emails`)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new UserActions.FetchedEmails(action.uid, data && data.EmailAddrs ? data.EmailAddrs : data)))
				);
			}
		}

		switch (action.constructor) {
		case UserActions.SubmitSignup:
			trackPromise(
				UserBackend.fetch(`/.api/join`, {
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
					.then((data) => {
						Dispatcher.Stores.dispatch(new UserActions.SignupCompleted(action.email, data));
						if (data.Success) {
							window.location.href = "/tour";
						}
					})
				);
			break;
		case UserActions.SubmitLogin:
			trackPromise(
				UserBackend.fetch(`/.api/login`, {
					method: "POST",
					body: JSON.stringify({
						Login: action.login,
						Password: action.password,
					}),
				})
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => Dispatcher.Stores.dispatch(new UserActions.LoginCompleted(data)))
			);
			break;
		case UserActions.SubmitLogout:
			trackPromise(
				UserBackend.fetch(`/.api/logout`, {
					method: "POST",
					body: JSON.stringify({}),
				})
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new UserActions.LogoutCompleted(data));
						// Redirect on logout.
						if (data.Success) {
							window.location.href = "/";
						}
					})
			);
			break;
		case UserActions.SubmitForgotPassword:
			trackPromise(
				UserBackend.fetch(`/.api/forgot`, {
					method: "POST",
					body: JSON.stringify({
						Email: action.email,
					}),
				})
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new UserActions.ForgotPasswordCompleted(data));
					})
				);
			break;
		case UserActions.SubmitResetPassword:
			trackPromise(
				UserBackend.fetch(`/.api/reset`, {
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
					.then((data) => {
						Dispatcher.Stores.dispatch(new UserActions.ResetPasswordCompleted(data));
					})
			);
			break;
		}
	},
};

Dispatcher.Backends.register(UserBackend.__onDispatch);

export default UserBackend;
