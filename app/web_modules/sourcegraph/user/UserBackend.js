import * as UserActions from "sourcegraph/user/UserActions";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const UserBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case UserActions.SubmitSignup:
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
				.catch((err) => {
					console.error(err);
					return {Error: true, err: err};
				})
				.then((data) => {
					Dispatcher.Stores.dispatch(new UserActions.SignupCompleted(data));
					if (!data.Error) {
						window.location.href = "/";
					}
				});
			break;
		case UserActions.SubmitLogin:
			UserBackend.fetch(`/.api/login`, {
				method: "POST",
				body: JSON.stringify({
					Login: action.login,
					Password: action.password,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => {
					console.error(err);
					return {Error: true, err: err};
				})
				.then((data) => {
					Dispatcher.Stores.dispatch(new UserActions.LoginCompleted(data));
					if (!data.Error) {
						window.location.href = "/";
					}
				});
			break;
		case UserActions.SubmitLogout:
			UserBackend.fetch(`/.api/logout`, {
				method: "POST",
				body: JSON.stringify({}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => {
					console.error(err);
					return {Error: true, err: err};
				})
				.then((data) => {
					Dispatcher.Stores.dispatch(new UserActions.LogoutCompleted(data));
					if (!data.Error) {
						window.location.href = "/";
					}
				});
			break;
		case UserActions.SubmitForgotPassword:
			UserBackend.fetch(`/.api/forgot`, {
				method: "POST",
				body: JSON.stringify({
					Email: action.email,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => {
					console.error(err);
					return {Error: true, err: err};
				})
				.then((data) => {
					Dispatcher.Stores.dispatch(new UserActions.ForgotPasswordCompleted(data));
					if (!data.Error) {
						window.location.href = "/";
					}
				});
			break;
		case UserActions.SubmitResetPassword:
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
				.catch((err) => {
					console.error(err);
					return {Error: true, err: err};
				})
				.then((data) => {
					Dispatcher.Stores.dispatch(new UserActions.ForgotPasswordCompleted(data));
					if (!data.Error) {
						window.location.href = "/";
					}
				});
			break;
		}
	},
};

Dispatcher.Backends.register(UserBackend.__onDispatch);

export default UserBackend;
