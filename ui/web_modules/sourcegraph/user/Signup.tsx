import * as React from "react";

import { PageTitle, Panel, SignupLoginAuth } from "sourcegraph/components";
import { layout } from "sourcegraph/components/utils";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import "sourcegraph/user/UserBackend"; // for side effects

export const SignupForm = (): JSX.Element => <SignupLoginAuth>
	To sign up, please authorize <br {...layout.hide.notSm } /> private code with GitHub:
</SignupLoginAuth>;

export const SignupComp = (): JSX.Element => <Panel hoverLevel="low" style={{ margin: "auto", maxWidth: "30rem" }}>
	<PageTitle title="Sign Up" />
	<SignupForm />
</Panel>;

export const Signup = redirectIfLoggedIn("/", SignupComp);
