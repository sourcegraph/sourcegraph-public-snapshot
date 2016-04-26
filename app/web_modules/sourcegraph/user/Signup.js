import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects
import redirectIfLoggedIn from "sourcegraph/user/redirectIfLoggedIn";
import CSSModules from "react-css-modules";
import style from "./styles/user.css";

class Signup extends Container {
	constructor(props) {
		super(props);
		this._loginInput = null;
		this._passwordInput = null;
		this._emailInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions.get("signup");
		state.authResponse = UserStore.authResponses.get("signup");
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		Dispatcher.Stores.dispatch(new UserActions.SubmitSignup());
		Dispatcher.Backends.dispatch(new UserActions.SubmitSignup(
			this._loginInput.value,
			this._passwordInput.value,
			this._emailInput.value,
		));
	}

	render() {
		return (
			<form styleName="container" onSubmit={this._handleSubmit}>
				<Helmet title="Sign Up" />
				<div styleName="title">Get started with Sourcegraph</div>
				<div styleName="action">
					<Input type="text"
						id="e2etest-login-field"
						placeholder="Username"
						domRef={(e) => this._loginInput = e}
						autoFocus={true}
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
				<div styleName="action">
					<Input type="email"
						id="e2etest-email-field"
						placeholder="Email"
						domRef={(e) => this._emailInput = e}
						block={true}
						required={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						id="e2etest-register-button"
						block={true}
						loading={this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)}>Create account & add GitHub repositories</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">{this.state.authResponse.Error.body.message}</div>
				}
				<div styleName="subtext">By creating an account you agree to our <a href="/privacy">privacy policy</a> and <a href="/legal">terms</a>.</div>
				<div styleName="alt-action">
					<span>Already have an account?</span>
					<span styleName="alt-button"><Link to="/login"><Button size="small" outline={true}>Sign in</Button></Link></span>
				</div>
			</form>
		);
	}
}

export default redirectIfLoggedIn("/", CSSModules(Signup, style));
