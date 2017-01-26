import { History } from "history";
import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { Component } from "sourcegraph/Component";
import { Heading } from "sourcegraph/components";
import { GitHubAuthButton } from "sourcegraph/components/GitHubAuthButton";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { whitespace } from "sourcegraph/components/utils";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import * as styles from "sourcegraph/user/styles/accountForm.css";
import "sourcegraph/user/UserBackend"; // for side effects

export interface PartialRouterLocation {
	pathname: string;
	hash: string;
}

function addQueryObjToURL(base: RouterLocation, urlOrPathname: string | PartialRouterLocation, queryObj: History.Query): RouterLocation {
	if (typeof urlOrPathname === "string") {
		urlOrPathname = { pathname: urlOrPathname } as RouterLocation;
	}
	return Object.assign({}, base, urlOrPathname, { query: queryObj });
}

interface Props {
	location: any;

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | RouterLocation;
	newUserReturnTo: PartialRouterLocation;
	queryObj: History.Query;
}

type State = any;

export class SignupForm extends Component<Props, State> {
	render(): JSX.Element | null {
		const newUserRedirLocation = addQueryObjToURL(this.props.location, this.props.newUserReturnTo, Object.assign({}, this.props.queryObj, { modal: "afterPrivateCodeSignup" }));
		return (
			<div>
				<div className={styles.form}>
					<Heading level={3} align="center" underline="orange">Get started with Sourcegraph</Heading>
					<GitHubAuthButton
						scopes="user:email"
						newUserReturnTo={newUserRedirLocation}
						returnTo={this.props.location}
						tabIndex={1}
						block={true}
						style={{ marginBottom: whitespace[2] }}
						secondaryText="Always free">Public code only</GitHubAuthButton>
					<GitHubAuthButton
						color="purple"
						newUserReturnTo={newUserRedirLocation}
						returnTo={this.props.location}
						tabIndex={2}
						block={true}
						secondaryText="14 days free">Private + public code</GitHubAuthButton>
					<p style={{ textAlign: "center" }}>
						By signing up, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
					</p>
					<p style={{ textAlign: "center" }}>
						Already have an account? <LocationStateToggleLink href="/login" modalName="login" location={location}>Log in.</LocationStateToggleLink>
					</p>
				</div>
			</div>
		);
	}
}

export const defaultOnboardingPath: PartialRouterLocation = {
	pathname: "/github.com/sourcegraph/checkup/-/blob/checkup.go",
	hash: "#L153",
};

function SignupComp(props: { location: any }): JSX.Element {
	return (
		<div className={styles.full_page}>
			<PageTitle title="Sign Up" />
			<SignupForm {...props}
				returnTo="/" queryObj={{ ob: "chrome" }}
				newUserReturnTo={defaultOnboardingPath} />
		</div>
	);
}

export const Signup = redirectIfLoggedIn("/", { ob: "chrome" }, SignupComp);
