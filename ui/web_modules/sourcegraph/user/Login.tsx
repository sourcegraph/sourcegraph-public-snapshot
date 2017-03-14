import * as React from "react";
import { Panel, SignupLoginAuth } from "sourcegraph/components";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";
import "sourcegraph/user/UserBackend"; // for side effects

// Login is the standalone login page.
export const LoginForm = (): JSX.Element => <Panel hoverLevel="low" style={{ margin: "auto", maxWidth: "30rem" }}>
	<PageTitle title="Sign In" />
	<SignupLoginAuth >
		To log in, authorize with GitHub:
	</SignupLoginAuth>
</Panel>;

export const Login = redirectIfLoggedIn("/", LoginForm);
