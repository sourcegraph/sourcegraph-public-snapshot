import React from "react";
import {Link} from "react-router";
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
		state.pendingAuthAction = UserStore.pendingAuthActions.get("reset");
		state.authResponse = UserStore.authResponses.get("reset");
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
						domRef={(e) => this._passwordInput = e}
						autoFocus={true}
						block={true}
						required={true} />
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="Confirm password"
						domRef={(e) => this._confirmInput = e}
						block={true}
						required={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction}>Reset Password</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">{this.state.authResponse.Error.body.message}</div>
				}
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Success &&
					<div styleName="success-text">Your password has been reset!<span styleName="alt-button"><Link to="/login"><Button size="small" outline={true}>Sign in</Button></Link></span></div>
				}
			</form>
		);
	}
}

export default CSSModules(ResetPassword, style);
