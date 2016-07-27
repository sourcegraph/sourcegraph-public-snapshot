import * as React from "react";
import Helmet from "react-helmet";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects
import redirectIfLoggedIn from "sourcegraph/user/redirectIfLoggedIn";
import CSSModules from "react-css-modules";
import style from "sourcegraph/user/styles/accountForm.css";

// TODO: prevent mounting this component if user is logged in
class ForgotPassword extends Container {
	constructor(props) {
		super(props);
		this._emailInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions.get("forgot");
		state.authResponse = UserStore.authResponses.get("forgot");
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		Dispatcher.Stores.dispatch(new UserActions.SubmitForgotPassword());
		Dispatcher.Backends.dispatch(new UserActions.SubmitForgotPassword(this._emailInput.value));
	}

	render() {
		return (
			<form styleName="full-page form" onSubmit={this._handleSubmit}>
				<Helmet title="Forgot Password" />
				<h1>Forgot your password?</h1>
				<label>
					<span>Email address</span>
					<Input type="email"
						placeholder="Email"
						domRef={(e) => this._emailInput = e}
						autoFocus={true}
						block={true}
						required={true} />
				</label>
				<Button color="blue"
					block={true}
					loading={this.state.pendingAuthAction}>Reset Password</Button>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="error">{this.state.authResponse.Error.body.message}</div>
				}
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Success &&
					<div styleName="success">Email sent - check your inbox!</div>
				}
			</form>
		);
	}
}

export default redirectIfLoggedIn("/", CSSModules(ForgotPassword, style, {allowMultiple: true}));
