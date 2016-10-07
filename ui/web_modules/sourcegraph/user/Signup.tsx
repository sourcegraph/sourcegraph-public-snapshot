import {Location} from "history";
import * as React from "react";
import Helmet from "react-helmet";
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
	queryObj: History.Query;
}

type State = any;

export class SignupForm extends Component<Props, State> {
	render(): JSX.Element | null {
		const redirQueryObj = Object.assign({}, this.props.location.query, this.props.queryObj);
		const redirRouteObj = typeof this.props.returnTo === "string" ? {pathname: this.props.returnTo} : this.props.returnTo;
		const redirLocation = Object.assign({}, this.props.location || null, redirRouteObj, {query: redirQueryObj});

		return (
			<div className={styles.form}>
				<Heading level="3" align="center" underline="orange">Welcome to Sourcegraph</Heading>
				<GitHubAuthButton returnTo={redirLocation} tabIndex={1} key="1" block={true}>Continue with GitHub</GitHubAuthButton>
				<p className={styles.mid_text}>
					By continuing with GitHub, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
				</p>
			</div>
		);
	}
}

function SignupComp(props: {location: any}): JSX.Element {
	return (
		<div className={styles.full_page}>
			<Helmet title="Sign Up" />
			<SignupForm {...props}
				returnTo="/" queryObj={{ob: "chrome"}} />
		</div>
	);
}

export const Signup = redirectIfLoggedIn("/", {ob: "chrome"}, SignupComp);
