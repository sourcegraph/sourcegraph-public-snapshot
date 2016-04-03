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
class ResetPassword extends Container {
	constructor(props) {
		super(props);
		this._passwordInput = null;
		this._confirmInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.token = state.location.query && state.location.query.token; // TODO: error handling (missing token)
		state.pendingAuthAction = UserStore.pendingAuthAction;
		state.authResponse = UserStore.authResponse;
	}

	stores() { return [UserStore]; }

	_handleSubmit() {
		Dispatcher.Stores.dispatch(new UserActions.SubmitResetPassword());
		Dispatcher.Backends.dispatch(new UserActions.SubmitResetPassword(
			this._passwordInput.getValue(),
			this._confirmInput.getValue(),
			this.state.token
		));
	}

	render() {
		return (
			<div styleName="container">
				<div styleName="title">Reset your password</div>
				<div styleName="subtext">
					Make sure to pick a good one!!
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="New password"
						ref={(c) => this._passwordInput = c}
						autoFocus={true}
						block={true} />
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="Confirm password"
						ref={(c) => this._confirmInput = c}
						block={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction}
						onClick={this._handleSubmit}>Reset Password</Button>
				</div>
			</div>
		);
	}
}

export default CSSModules(ResetPassword, style);
