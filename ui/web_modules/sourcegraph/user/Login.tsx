// tslint:disable

import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import {Button, Input, Heading} from "sourcegraph/components/index";
import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";
import "sourcegraph/user/UserBackend"; // for side effects
import redirectIfLoggedIn from "sourcegraph/user/redirectIfLoggedIn";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import * as styles from "sourcegraph/user/styles/accountForm.css";

type Props = {
	onLoginSuccess: () => void,
	location: any,

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | HistoryModule.Location,
};

export class LoginForm extends Container<Props, any> {
	_loginInput: any;
	_passwordInput: any;

	constructor(props) {
		super(props);
		this._loginInput = null;
		this._passwordInput = null;
		this._handleSubmit = this._handleSubmit.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions["login"] || false;
		state.authResponse = UserStore.authResponses["login"] || null;

		// These are set by the GitHub OAuth2 receive endpoint if there is an
		// error.
		state.githubError = (props.location.query && props.location.query["github-login-error"]) || null;
	}

	onStateTransition(prevState, nextState) {
		if (prevState.authResponse !== nextState.authResponse) {
			if (nextState.submitted && nextState.authResponse && nextState.authResponse.Success) {
				setTimeout(() => this.props.onLoginSuccess());
			}
		}
	}

	stores() { return [UserStore]; }

	_handleSubmit(ev) {
		ev.preventDefault();
		this.setState({submitted: true}, () => {
			let action = new UserActions.SubmitLogin(
				this._loginInput.value,
				this._passwordInput.value,
			);
			Dispatcher.Stores.dispatch(action);
			Dispatcher.Backends.dispatch(action);
		});
	}

	render(): JSX.Element | null {
		return (
			<form {...this.props} onSubmit={this._handleSubmit} className={styles.form}>
				<Heading level="3" align="center" underline="orange">Sign in to Sourcegraph</Heading>
				{this.state.githubError && <div className={styles.error}>Sorry, signing in via GitHub didn't work. (Check your organization's GitHub 3rd-party application settings.) Try <Link to="/join?github-error=from-login">creating a separate Sourcegraph account</Link>.</div>}
				<GitHubAuthButton returnTo={this.state.returnTo} tabIndex="1" block={true}>Continue with GitHub</GitHubAuthButton>
				<p className={styles.divider}>or</p>
				<label>
					<span>Username</span>
					<Input type="text"
						id="e2etest-login-field"
						name="username"
						autoComplete="username"
						autoFocus={true}
						autoCapitalize={false}
						autoCorrect={false}
						minLength="3"
						tabIndex="2"
						domRef={(e) => this._loginInput = e}
						block={true}
						required={true} />
				</label>
				<label>
					<span>Password</span>
					<Link className={styles.label_link} to="/forgot">Forgot password?</Link>
					<Input type="password"
						id="e2etest-password-field"
						autoComplete="current-password"
						name="password"
						tabIndex="3"
						domRef={(e) => this._passwordInput = e}
						block={true}
						required={true} />
				</label>
				<Button color="normal"
					id="e2etest-login-button"
					tabIndex="4"
					block={true}
					loading={this.state.submitted && (this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error))}>Sign in</Button>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div className={styles.error}>{this.state.authResponse.Error.body.message}</div>
				}
				<p className={styles.sub_text}>
					No account yet? <Link tabIndex="5" to="/join">Sign up.</Link>
				</p>
			</form>
		);
	}
}
let StyledLoginForm = LoginForm;

// Login is the standalone login page.
function Login(props: any, {router}) {
	return (
		<div className={styles.full_page}>
			<Helmet title="Sign In" />
			<StyledLoginForm {...props}
				returnTo="/"
				onLoginSuccess={() => router.replace("/")} />
		</div>
	);
}
(Login as any).contextTypes = {
	router: React.PropTypes.object.isRequired,
};


export default redirectIfLoggedIn("/", Login);
