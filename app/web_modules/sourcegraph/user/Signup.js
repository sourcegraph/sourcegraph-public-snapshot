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
import style from "sourcegraph/user/styles/accountForm.css";

export class SignupForm extends Container {
	static propTypes = {
		onSignupSuccess: React.PropTypes.func.isRequired,
	};
	state = {
		submitted: false,
	};

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


	onStateTransition(prevState, nextState) {
		if (prevState.authResponse !== nextState.authResponse) {
			if (nextState.submitted && nextState.authResponse && nextState.authResponse.Success) {
				setTimeout(() => this.props.onSignupSuccess());
			}
		}
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		this.setState({submitted: true}, () => {
			Dispatcher.Stores.dispatch(new UserActions.SubmitSignup());
			Dispatcher.Backends.dispatch(new UserActions.SubmitSignup(
				this._loginInput.value,
				this._passwordInput.value,
				this._emailInput.value,
			));
		});
	}

	render() {
		return (
			<form {...this.props} onSubmit={this._handleSubmit} styleName="form">
				<div styleName="title">Sign up for Sourcegraph</div>
				<label>
					<span>Username</span>
					<Input type="text"
						id="e2etest-login-field"
						name="username"
						domRef={(e) => this._loginInput = e}
						autoComplete="username"
						autoFocus={true}
						autoCapitalize={false}
						autoCorrect={false}
						minLength="3"
						tabIndex="1"
						block={true}
						required={true} />
				</label>
				<label>
					<span>Email address</span>
					<Input type="email"
						id="e2etest-email-field"
						name="email"
						autoComplete="email"
						autoCapitalize={false}
						tabIndex="2"
						domRef={(e) => this._emailInput = e}
						block={true}
						required={true} />
				</label>
				<label>
					<span>Password</span>
					<Input type="password"
						id="e2etest-password-field"
						name="password"
						autoComplete="new-password"
						domRef={(e) => this._passwordInput = e}
						tabIndex="3"
						block={true}
						required={true} />
				</label>
				<p styleName="mid-text">
					By creating an account, you agree to our <a href="/privacy">privacy policy</a> and <a href="/legal">terms</a>.
				</p>
				<Button color="primary"
					id="e2etest-register-button"
					tabIndex="4"
					block={true}
					loading={this.state.submitted && (this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error))}>Create account & add GitHub repositories</Button>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="error">{this.state.authResponse.Error.body.message}</div>
				}
				<p styleName="sub-text">
					Already have an account? <Link to="/login">Sign in.</Link>
				</p>
			</form>
		);
	}
}
SignupForm = CSSModules(SignupForm, style);

function Signup(props, {router}) {
	return (
		<div styleName="full-page">
			<Helmet title="Sign Up" />
			<SignupForm {...props}
				onSignupSuccess={() => router.replace("/")} />
		</div>
	);
}
Signup.contextTypes = {
	router: React.PropTypes.object.isRequired,
};

export default redirectIfLoggedIn("/", CSSModules(Signup, style));
