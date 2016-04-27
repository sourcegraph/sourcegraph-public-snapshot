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

export class LoginForm extends Container {
	static propTypes = {
		onLoginSuccess: React.PropTypes.func.isRequired,
	};

	state = {
		submitted: false,
	};

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
			Dispatcher.Stores.dispatch(new UserActions.SubmitLogin());
			Dispatcher.Backends.dispatch(new UserActions.SubmitLogin(
				this._loginInput.value,
				this._passwordInput.value,
			));
		});
	}

	render() {
		return (
			<form {...this.props} onSubmit={this._handleSubmit} styleName="form">
				<h1 styleName="title">Sign in to Sourcegraph</h1>
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
						tabIndex="1"
						domRef={(e) => this._loginInput = e}
						block={true}
						required={true} />
				</label>
				<label>
					<span>Password</span>
					<Link styleName="label-link" to="/forgot">Forgot password?</Link>
					<Input type="password"
						id="e2etest-password-field"
						autoComplete="current-password"
						name="password"
						tabIndex="2"
						domRef={(e) => this._passwordInput = e}
						block={true}
						required={true} />
				</label>
				<Button color="primary"
					id="e2etest-login-button"
					tabIndex="3"
					block={true}
					loading={this.state.submitted && (this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error))}>Sign in</Button>
				{!this.state.pendingAuthAction && this.state.authResponse && this.state.authResponse.Error &&
					<div styleName="error">{this.state.authResponse.Error.body.message}</div>
				}
				<p styleName="sub-text">
					No account yet? <Link to="/join">Sign up.</Link>
				</p>
			</form>
		);
	}
}
LoginForm = CSSModules(LoginForm, style);

// Login is the standalone login page.
function Login(props, {router}) {
	return (
		<div styleName="full-page">
			<Helmet title="Sign In" />
			<LoginForm {...props}
				onLoginSuccess={() => router.replace("/")} />
		</div>
	);
}
Login.contextTypes = {
	router: React.PropTypes.object.isRequired,
};


export default redirectIfLoggedIn("/", CSSModules(Login, style));
