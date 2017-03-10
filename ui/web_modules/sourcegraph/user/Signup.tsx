import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { PageTitle, Panel, SignupLoginAuth } from "sourcegraph/components";
import { layout } from "sourcegraph/components/utils";
import { defaultOnboardingPath } from "sourcegraph/user/Auth";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	// returnTo is where the user should be redirected after an OAuth login flow,
	// either a URL path or a Location object.
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
}

export const SignupForm = (props: Props): JSX.Element => <SignupLoginAuth {...props}>
	To sign up, please authorize <br {...layout.hide.notSm } /> private code with GitHub:
</SignupLoginAuth>;

export const SignupComp = (): JSX.Element => <Panel hoverLevel="low" style={{ margin: "auto", maxWidth: "30rem" }}>
	<PageTitle title="Sign Up" />
	<SignupForm returnTo="/" newUserReturnTo={defaultOnboardingPath} />
</Panel>;

export const Signup = redirectIfLoggedIn("/", SignupComp);
