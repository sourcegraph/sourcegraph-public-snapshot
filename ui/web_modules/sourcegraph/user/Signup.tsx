import { History } from "history";
import * as React from "react";

import { context } from "sourcegraph/app/context";
import { Router } from "sourcegraph/app/router";
import { RouterLocation } from "sourcegraph/app/router";
import { GitHubAuthButton } from "sourcegraph/components/GitHubAuthButton";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { whitespace } from "sourcegraph/components/utils";
import { LoggableEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import "sourcegraph/user/UserBackend"; // for side effects
import { oauthProvider, urlToOAuth } from "sourcegraph/util/urlTo";

export interface PartialRouterLocation {
	pathname: string;
	hash: string;
}

export function addQueryObjToURL(base: RouterLocation, urlOrPathname: string | PartialRouterLocation, queryObj: History.Query): RouterLocation {
	if (typeof urlOrPathname === "string") {
		urlOrPathname = { pathname: urlOrPathname } as RouterLocation;
	}
	return Object.assign({}, base, urlOrPathname, { query: { modal: "afterSignup", ...queryObj } });
}

interface Props {
	location: any;

	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo: string | RouterLocation;
	newUserReturnTo: PartialRouterLocation;
}

export function SignupForm(props: Props): JSX.Element {
	const publicNewUserRedir = addQueryObjToURL(props.location, props.newUserReturnTo, {});
	const privateNewUserRedir = addQueryObjToURL(props.location, props.newUserReturnTo, { private: true });
	return <div>
		<GitHubAuthButton
			scope="public"
			newUserReturnTo={publicNewUserRedir}
			returnTo={props.returnTo}
			tabIndex={1}
			block={true}
			style={{ marginBottom: whitespace[2] }}
			secondaryText="Always free">Public code only</GitHubAuthButton>
		<GitHubAuthButton
			scope="private"
			color="purple"
			newUserReturnTo={privateNewUserRedir}
			returnTo={props.returnTo}
			tabIndex={2}
			block={true}
			secondaryText="14 days free">Private + public code</GitHubAuthButton>
		<p style={{ textAlign: "center" }}>
			By signing up, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
		</p>
		<p style={{ textAlign: "center" }}>
			Already have an account? <LocationStateToggleLink href="/login" modalName="login" location={location}>Log in.</LocationStateToggleLink>
		</p>
	</div>;
}

export const defaultOnboardingPath: PartialRouterLocation = {
	pathname: "/github.com/sourcegraph/checkup/-/blob/checkup.go",
	hash: "#L153",
};

function SignupComp(props: { location: any }): JSX.Element {
	return (
		<div style={{ margin: "auto", maxWidth: "30rem" }}>
			<PageTitle title="Sign Up" />
			<SignupForm {...props} returnTo="/" newUserReturnTo={defaultOnboardingPath} />
		</div>
	);
}

export const Signup = redirectIfLoggedIn("/", SignupComp);

export function ghCodeAction(router: Router, privateCode: boolean): ActionForm {
	const newUserPath = router.location.pathname.indexOf("/-/blob/") !== -1 ? { pathname: router.location.pathname, hash: router.location.hash } : defaultOnboardingPath;
	const base = Object.assign(router.location, { pathname: "" });
	const newUserReturnTo = addQueryObjToURL(base, newUserPath, { private: privateCode });
	return getAuthAction({
		eventObject: Events.OAuth2FlowGitHub_Initiated,
		provider: "github",
		scopes: privateCode ? "read:org,user:email,repo" : "read:org,user:email",
		returnTo: router.location,
		newUserReturnTo,
	});
}

/**
 * An action form contains an JSX element that must be included in the DOM and
 * an action to submit that form.
 */
interface ActionForm {
	submit: () => void;
	form: JSX.Element;
}

export interface AuthProps {
	eventObject: LoggableEvent;
	pageName?: string;

	provider: oauthProvider;
	scopes: string;
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
}

/**
 * Get an authorization action and form.
 */
export function getAuthAction(props: AuthProps): ActionForm {
	let url = urlToOAuth(
		props.provider,
		props.scopes,
		props.returnTo || null,
		props.newUserReturnTo || null,
	);

	let authForm: HTMLFormElement | null = null;
	const submitAuthForm = () => {
		if (authForm) {
			authForm.submit();
		}
	};
	const logEvent = () => {
		props.eventObject.logEvent({ page_name: props.pageName || "" });
	};

	return {
		submit: submitAuthForm,
		form: <form
			action={url}
			method="POST"
			onSubmit={logEvent}
			ref={el => authForm = el}
			style={{ display: "none" }} >
			<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
		</form>
	};
}
