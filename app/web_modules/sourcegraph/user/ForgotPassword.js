import React from "react";
import Helmet from "react-helmet";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects

import CSSModules from "react-css-modules";
import style from "./styles/user.css";

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
			<form styleName="container" onSubmit={this._handleSubmit}>
				<Helmet title="Forgot Password" />
				<div styleName="title">Forgot your password?</div>
				<div styleName="subtext">
					It happens to the best of us.
					<br />
					Enter your email below and we'll send you a link to recover your password.
				</div>
				<div styleName="action">
					<Input type="email"
						placeholder="Email"
						domRef={(e) => this._emailInput = e}
						autoFocus={true}
						block={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction}>Reset Password</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">Sorry, there's been a problem.<br />{this.state.authResponse.Error.message}</div>
				}
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Success &&
					<div styleName="success-text">Email sent - check your inbox!</div>
				}
			</form>
		);
	}
}

export default CSSModules(ForgotPassword, style);
