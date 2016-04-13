import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as UserActions from "sourcegraph/user/UserActions";

export class UserStore extends Store {
	reset() {
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
	}

	__onDispatch(action) {
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
