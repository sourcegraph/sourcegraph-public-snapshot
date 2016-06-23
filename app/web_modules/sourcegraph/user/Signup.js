import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input, Heading} from "sourcegraph/components";

import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import "sourcegraph/user/UserBackend"; // for side effects
import redirectIfLoggedIn from "sourcegraph/user/redirectIfLoggedIn";
import CSSModules from "react-css-modules";
import style from "sourcegraph/user/styles/accountForm.css";

export class SignupForm extends Container {
	static propTypes = {
		onSignupSuccess: React.PropTypes.func.isRequired,
		location: React.PropTypes.object.isRequired,

		// returnTo is where the user should be redirected after an OAuth login flow,
		// either a URL path or a Location object.
		returnTo: React.PropTypes.oneOfType(React.PropTypes.string, React.PropTypes.object).isRequired,
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

		// These are set by the GitHub OAuth2 receive endpoint if there is an
		// error.
		state.githubError = (props.location.query && props.location.query["github-signup-error"]) || null;
		state.githubLogin = (props.location.query && props.location.query.login) || null;
		state.githubEmail = (props.location.query && props.location.query.email) || null;
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
				<Heading level="3" align="center" underline="orange">Sign up for Sourcegraph</Heading>
				{!this.state.githubError && [
					<GitHubAuthButton returnTo={this.state.returnTo} tabIndex="1" key="1" block={true}>Continue with GitHub</GitHubAuthButton>,
					<p key="2" styleName="divider">or</p>,
				]}
				{this.state.githubError === "username-or-email-taken" && <div styleName="error">Your GitHub username <strong>{this.state.githubLogin}</strong> {this.state.githubEmail && <span>or email <strong>{this.state.githubEmail}</strong></span>} is already taken on Sourcegraph. Sign up on Sourcegraph with a different username/email, then link your GitHub account again.</div>}
				{this.state.githubError === "unknown" && <div styleName="error">Sorry, signing up via GitHub didn't work. (Check your organization's GitHub 3rd-party application settings.) Try creating a separate Sourcegraph account below.</div>}
				<label>
					<span>Username</span>
					<Input type="text"
						id="e2etest-login-field"
						name="username"
						defaultValue={this.state.githubLogin || null}
						domRef={(e) => this._loginInput = e}
						autoComplete="username"
						autoFocus={true}
						autoCapitalize={false}
						autoCorrect={false}
						minLength="3"
						tabIndex="2"
						block={true}
						required={true} />
				</label>
				<label>
					<span>Email address</span>
					<Input type="email"
						id="e2etest-email-field"
						name="email"
						defaultValue={this.state.githubEmail || null}
						autoComplete="email"
						autoCapitalize={false}
						tabIndex="3"
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
						tabIndex="4"
						block={true}
						required={true} />
				</label>
				<p styleName="mid-text">
					By creating an account, you agree to our <a href="/-/privacy" target="_blank">privacy policy</a> and <a href="/-/terms" target="_blank">terms</a>.
				</p>
				<Button
					color={this.state.githubError ? "blue" : "default"}
					id="e2etest-register-button"
					tabIndex="5"
					block={true}
					loading={this.state.submitted && (this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error))}>Create account</Button>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="error">{this.state.authResponse.Error.body.message}</div>
				}
				<p styleName="sub-text">
					Already have an account? <Link tabIndex="6" to="/login">Sign in.</Link>
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
				returnTo="/"
				onSignupSuccess={() => router.replace({...location, state: {...location.state, _onboarding: "new-user"}})} />
		</div>
	);
}
Signup.contextTypes = {
	router: React.PropTypes.object.isRequired,
};

export default redirectIfLoggedIn("/", CSSModules(Signup, style));
