import React from "react";
import {Link} from "react-router";

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
				<div styleName="title">Get started with Sourcegraph</div>
				<div styleName="action">
					<Input type="text"
						placeholder="Username"
						domRef={(e) => this._loginInput = e}
						autoFocus={true}
						block={true} />
				</div>
				<div styleName="action">
					<Input type="password"
						placeholder="Password"
						domRef={(e) => this._passwordInput = e}
						block={true} />
				</div>
				<div styleName="action">
					<Input type="email"
						placeholder="Email"
						domRef={(e) => this._emailInput = e}
						block={true} />
				</div>
				<div styleName="button">
					<Button color="primary"
						block={true}
						loading={this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)}>Create Account</Button>
				</div>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="errtext">Sorry, there's been a problem.<br />{this.state.authResponse.Error.message}</div>
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

export default CSSModules(Signup, style);
