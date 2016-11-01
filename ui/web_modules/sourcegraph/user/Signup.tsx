import {History} from "history";
import * as React from "react";
import Helmet from "react-helmet";
import {context} from "sourcegraph/app/context";
import {Component} from "sourcegraph/Component";
import {Heading} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {Location} from "sourcegraph/Location";
import {redirectIfLoggedIn} from "sourcegraph/user/redirectIfLoggedIn";
import * as styles from "sourcegraph/user/styles/accountForm.css";
import "sourcegraph/user/UserBackend"; // for side effects
import {urlToOAuth} from "sourcegraph/util/urlTo";

interface Props {
	location: any;

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | Location;
	queryObj: History.Query;
}

type State = any;

export class SignupForm extends Component<Props, State> {
	_submitForm(): void {
		let form = document.getElementById("form");
		if (form) {
			(form as HTMLFormElement).submit();
		}
	}

	render(): JSX.Element | null {
		const redirQueryObj = Object.assign({}, this.props.location.query, this.props.queryObj);
		const redirRouteObj = typeof this.props.returnTo === "string" ? {pathname: this.props.returnTo} : this.props.returnTo;
		const redirLocation = Object.assign({}, this.props.location || null, redirRouteObj, {query: redirQueryObj});

		const publicCodeURL = urlToOAuth("github", "read:org,user:email", this.props.returnTo || null);
		return (
			<div className={styles.form}>
				<Heading level={3} align="center" underline="orange">Welcome to Sourcegraph</Heading>
				<GitHubAuthButton returnTo={redirLocation} tabIndex={1} key="1" block={true}>Continue with GitHub</GitHubAuthButton>
				<form id="form" method="POST" action={publicCodeURL}>
					<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
					<p className={styles.mid_text}>
						Or <a onClick={this._submitForm.bind(this)}>sign in</a> to view public code <i>only</i>. By continuing with GitHub, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
					</p>
				</form>
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
