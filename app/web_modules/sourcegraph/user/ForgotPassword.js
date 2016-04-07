import React from "react";

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
		state.pendingAuthAction = UserStore.pendingAuthAction;
		state.authResponse = UserStore.authResponse;
	}

	stores() { return [UserStore]; }

	_handleSubmit() {
		Dispatcher.Stores.dispatch(new UserActions.SubmitForgotPassword());
		Dispatcher.Backends.dispatch(new UserActions.SubmitForgotPassword(this._emailInput.getValue()));
	}

	render() {
		return (
			<div styleName="container">
				<div styleName="title">Forgot your password?</div>
				<div styleName="subtext">
					It happens to the best of us.
					<br />
					Enter your email below and we'll send you a link to recover your password.
				</div>
				<div styleName="action">
					<Input type="email"
						placeholder="Email"
						ref={(c) => this._emailInput = c}
						onSubmit={this._handleSubmit}
						autoFocus={true}
						block={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)}
						onClick={this._handleSubmit}>Reset Password</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">Sorry, there's been a problem.<br />{this.state.authResponse.err.message}</div>
				}
			</div>
		);
	}
}

export default CSSModules(ForgotPassword, style);
