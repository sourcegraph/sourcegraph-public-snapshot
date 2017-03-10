import * as React from "react";
import { RouterLocation } from "sourcegraph/app/router";
import { Panel, SignupLoginAuth } from "sourcegraph/components";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo?: string | RouterLocation;
};

export const LoginForm = (props: Props): JSX.Element => <SignupLoginAuth {...props}>
	To log in, authorize with GitHub:
</SignupLoginAuth>;

// Login is the standalone login page.
export const LoginComp = (): JSX.Element => <Panel hoverLevel="low" style={{ margin: "auto", maxWidth: "30rem" }}>
	<PageTitle title="Sign In" />
	<LoginForm returnTo="/" />
</Panel>;

export const Login = redirectIfLoggedIn("/", LoginComp);
