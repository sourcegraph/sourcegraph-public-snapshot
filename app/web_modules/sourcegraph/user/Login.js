import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects

import CSSModules from "react-css-modules";
import style from "./styles/user.css";

class Login extends Container {
	constructor(props) {
		super(props);
		this._loginInput = null;
		this._passwordInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions.get("login");
		state.authResponse = UserStore.authResponses.get("login");
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		Dispatcher.Stores.dispatch(new UserActions.SubmitLogin());
		Dispatcher.Backends.dispatch(new UserActions.SubmitLogin(
				this._loginInput.value,
				this._passwordInput.value,
		));
	}

	render() {
		return (
			<form styleName="container" onSubmit={this._handleSubmit}>
				<Helmet title="Sign In" />
				<div styleName="title">Hey there, welcome back!</div>
				<div styleName="action">
					<Input type="text"
						id="e2etest-login-field"
						autoFocus={true}
						placeholder="Username"
						domRef={(e) => this._loginInput = e}
						block={true}
						required={true} />
				</div>
				<div styleName="action">
					<Input type="password"
						id="e2etest-password-field"
						placeholder="Password"
						domRef={(e) => this._passwordInput = e}
						block={true}
						required={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						id="e2etest-login-button"
						block={true}
						loading={this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)}>Sign in</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">{this.state.authResponse.Error.body.message}</div>
				}
				<div styleName="subtext">Oh no, <a href="/forgot">I forgot my password</a>.</div>
				<div styleName="alt-action">
					<span>Don't have an account yet?</span>
					<span styleName="alt-button"><Link to="/join"><Button size="small" outline={true}>Sign up</Button></Link></span>
				</div>
			</form>
		);
	}
}

export default CSSModules(Login, style);
