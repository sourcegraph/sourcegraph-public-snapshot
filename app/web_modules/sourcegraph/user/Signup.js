import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects

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
		state.pendingAuthAction = UserStore.pendingAuthAction;
		state.authResponse = UserStore.authResponse;
	}

	stores() { return [UserStore]; }

	_handleSubmit() {
		Dispatcher.Stores.dispatch(new UserActions.SubmitSignup());
		Dispatcher.Backends.dispatch(new UserActions.SubmitSignup(
			this._loginInput.getValue(),
			this._passwordInput.getValue(),
			this._emailInput.getValue()
		));
	}

	render() {
		return (
			<div styleName="container">
				<div styleName="title">Get started with Sourcegraph</div>
				<div styleName="action">
					<Input type="text"
						placeholder="Username"
						ref={(c) => this._loginInput = c}
						autoFocus={true}
						block={true} />
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="Password"
						ref={(c) => this._passwordInput = c}
						block={true} />
				</div>
				<div styleName="action">
					<Input type="email"
						placeholder="Email"
						ref={(c) => this._emailInput = c}
						block={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction}
						onClick={this._handleSubmit}>Create Account</Button>
				</div>
				<div styleName="subtext">By creating an account you agree to our <a href="/privacy">privacy policy</a> and <a href="/legal">terms</a>.</div>
				<div styleName="alt-action">
					<span>Already have an account?</span>
					<span styleName="alt-button">
						<Button color="default" outline={true} small={true}><a styleName="alt-link" href="/login">Sign in</a></Button>
					</span>
				</div>
			</div>
		);
	}
}

export default CSSModules(Signup, style);
