import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as UserActions from "sourcegraph/user/UserActions";

export class UserStore extends Store {
	reset() {
		this.pendingAuthAction = false;
		this.authResponse = null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case UserActions.SubmitSignup:
		case UserActions.SubmitLogin:
		case UserActions.SubmitLogout:
		case UserActions.SubmitForgotPassword:
		case UserActions.SubmitResetPassword:
			this.pendingAuthAction = true;
			break;
		case UserActions.SignupCompleted:
		case UserActions.LoginCompleted:
		case UserActions.LogoutCompleted:
		case UserActions.ForgotPasswordCompleted:
		case UserActions.ResetPasswordCompleted:
			this.pendingAuthAction = false;
			this.authResponse = deepFreeze(action.resp);
			break;
		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new UserStore(Dispatcher.Stores);
