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

	_handleSubmit(ev) {
		ev.preventDefault();
		Dispatcher.Stores.dispatch(new UserActions.SubmitResetPassword());
		Dispatcher.Backends.dispatch(new UserActions.SubmitResetPassword(
			this._passwordInput.value,
			this._confirmInput.value,
			this.state.token
		));
	}

	render() {
		return (
			<form styleName="container" onSubmit={this._handleSubmit}>
				<Helmet title="Reset Password" />
				<div styleName="title">Reset your password</div>
				<div styleName="subtext">
					Make sure to pick a good one!!
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="New password"
						ref={(e) => this._passwordInput = e}
						autoFocus={true}
						block={true} />
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="Confirm password"
						ref={(e) => this._confirmInput = e}
						block={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)}>Reset Password</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">Sorry, there's been a problem.<br />{this.state.authResponse.Error.message}</div>
				}
			</form>
		);
	}
}

export default CSSModules(ResetPassword, style);
