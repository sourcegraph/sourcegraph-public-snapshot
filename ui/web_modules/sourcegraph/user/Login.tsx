import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";
import {Heading} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {Location} from "sourcegraph/Location";
import {redirectIfLoggedIn} from "sourcegraph/user/redirectIfLoggedIn";
import * as styles from "sourcegraph/user/styles/accountForm.css";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	location: any;

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | Location;
};

export function LoginForm(props: Props): JSX.Element {
		return (
			<div className={styles.form}>
				<Heading level={3} align="center" underline="orange">Sign in to Sourcegraph</Heading>
				<GitHubAuthButton returnTo={props.returnTo || props.location} tabIndex={1} block={true}>Continue with GitHub</GitHubAuthButton>
				<p className={styles.sub_text}>
					No account yet? <Link tabIndex={5} to="/join">Sign up.</Link>
				</p>
				<p className={styles.mid_text}>
					By creating an account, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
				</p>
			</div>
		);
}

// Login is the standalone login page.
function LoginComp(props: {location: any}): JSX.Element | null {
	return (
		<div className={styles.full_page}>
			<Helmet title="Sign In" />
			<LoginForm location={props.location} returnTo="/" />
		</div>
	);
}

export const Login = redirectIfLoggedIn("/", {}, LoginComp);
