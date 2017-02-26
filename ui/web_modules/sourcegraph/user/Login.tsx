import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { Heading } from "sourcegraph/components";
import { GitHubAuthButton } from "sourcegraph/components/GitHubAuthButton";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import { addQueryObjToURL, defaultOnboardingPath } from "sourcegraph/user/Signup";
import * as styles from "sourcegraph/user/styles/accountForm.css";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	location: any;

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | RouterLocation;
};

export function LoginForm(props: Props): JSX.Element {
	// TODO(john): provide route pattern in `location` and use `RouterLocation` type.
	let newUserPath = props.location.pathname.indexOf("/-/blob/") !== -1 ? { pathname: props.location.pathname, hash: props.location.hash } : defaultOnboardingPath;
	const publicNewUserRedir = addQueryObjToURL(props.location, newUserPath, {});
	const privateNewUserRedir = addQueryObjToURL(props.location, newUserPath, { modal: "afterPrivateCodeSignup" });
	return (
		<div className={styles.form}>
			<Heading level={3} align="center" underline="orange">Sign in to Sourcegraph</Heading>
			<GitHubAuthButton scope="public" returnTo={props.returnTo || props.location} newUserReturnTo={publicNewUserRedir} tabIndex={1} block={true}>Public code only</GitHubAuthButton>
			<GitHubAuthButton scope="private" color="purple" returnTo={props.returnTo || props.location} newUserReturnTo={privateNewUserRedir} tabIndex={2} block={true}>Private + public code</GitHubAuthButton>
			<p style={{ textAlign: "center" }}>
				No account yet? <LocationStateToggleLink href="/join" modalName="join" location={location}>Sign up.</LocationStateToggleLink>
			</p>
			<p style={{ textAlign: "center" }}>
				By signing in, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
			</p>
		</div>
	);
}

// Login is the standalone login page.
function LoginComp(props: { location: any }): JSX.Element {
	return (
		<div className={styles.full_page}>
			<PageTitle title="Sign In" />
			<LoginForm location={props.location} returnTo="/" />
		</div>
	);
}

export const Login = redirectIfLoggedIn("/", LoginComp);
