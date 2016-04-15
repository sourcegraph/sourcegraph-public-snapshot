import * as UserActions from "sourcegraph/user/UserActions";
import Dispatcher from "sourcegraph/Dispatcher";
import EventLogger, {EventLocation} from "sourcegraph/util/EventLogger";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";

const UserBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
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
						if (!data.Error) {
							// Currently the only way to signup is from the signup page. When this changes we will
							// want to modify how we are tracking this event.
							EventLogger.logEventForPage("SubmitLinkGitHub", EventLocation.Signup);
							window.location.href = urlToGitHubOAuth;
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
					.then((data) => {
						Dispatcher.Stores.dispatch(new UserActions.LoginCompleted(data));
						if (!data.Error) {
							window.location.href = "/";
						}
					})
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
						if (!data.Error) {
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
