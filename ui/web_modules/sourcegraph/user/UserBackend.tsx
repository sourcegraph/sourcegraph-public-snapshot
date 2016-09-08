import * as Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

class UserBackendClass {
	fetch: any;

	constructor() {
		this.fetch = defaultFetch;
	}

	__onDispatch(payload: UserActions.Action): void {
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
					if (data.Success) { window.location.href = "/?ob=chrome"; }
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
					if (data.Success) { window.location.reload(); }
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
