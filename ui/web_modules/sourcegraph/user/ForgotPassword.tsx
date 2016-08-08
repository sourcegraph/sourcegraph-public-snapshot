// tslint:disable

import * as React from "react";
import Helmet from "react-helmet";
import * as classNames from "classnames";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input} from "sourcegraph/components/index";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";

import "sourcegraph/user/UserBackend"; // for side effects
import redirectIfLoggedIn from "sourcegraph/user/redirectIfLoggedIn";
import * as styles from "sourcegraph/user/styles/accountForm.css";

// TODO: prevent mounting this component if user is logged in
class ForgotPassword extends Container<any, any> {
	_emailInput: any;

	constructor(props) {
		super(props);
		this._emailInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions["forgot"] || false;
		state.authResponse = UserStore.authResponses["forgot"] || null;
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		let action = new UserActions.SubmitForgotPassword(this._emailInput.value);
		Dispatcher.Stores.dispatch(action);
		Dispatcher.Backends.dispatch(action);
	}

	render(): JSX.Element | null {
		return (
			<form className={classNames(styles.full_page, styles.form)} onSubmit={this._handleSubmit}>
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
					<div className={styles.error}>{this.state.authResponse.Error.body.message}</div>
				}
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Success &&
					<div className={styles.success}>Email sent - check your inbox!</div>
				}
			</form>
		);
	}
}

export default redirectIfLoggedIn("/", ForgotPassword);
