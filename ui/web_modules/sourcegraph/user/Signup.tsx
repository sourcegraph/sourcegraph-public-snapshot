import {Location} from "history";
import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";
import {Component} from "sourcegraph/Component";
import {Heading} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {redirectIfLoggedIn} from "sourcegraph/user/redirectIfLoggedIn";
import * as styles from "sourcegraph/user/styles/accountForm.css";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	location: any;

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | Location;
}

type State = any;

export class SignupForm extends Component<Props, State> {
	render(): JSX.Element | null {
		return (
			<div className={styles.form}>
				<Heading level="3" align="center" underline="orange">Sign up for Sourcegraph</Heading>
				<GitHubAuthButton returnTo={this.state.returnTo || this.props.location} tabIndex={1} key="1" block={true}>Continue with GitHub</GitHubAuthButton>
				<p className={styles.sub_text}>
					Already have an account? <Link tabIndex={6} to="/login">Sign in.</Link>
				</p>
				<p className={styles.mid_text}>
					By creating an account, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
				</p>
			</div>
		);
	}
}
let StyledSignupForm = SignupForm;

function SignupComp(props: {location: any}): JSX.Element {
	return (
		<div className={styles.full_page}>
			<Helmet title="Sign Up" />
			<StyledSignupForm {...props}
				returnTo="/" />
		</div>
	);
}

export const Signup = redirectIfLoggedIn("/", SignupComp);
