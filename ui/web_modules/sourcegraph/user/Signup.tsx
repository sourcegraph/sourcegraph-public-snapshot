import * as React from "react";

import { PageTitle, Panel, SignupLoginAuth } from "sourcegraph/components";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { layout } from "sourcegraph/components/utils";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";

export const SignupForm = (): JSX.Element => <SignupLoginAuth>
	To sign up, please authorize <br {...layout.hide.notSm } /> code with GitHub:
</SignupLoginAuth>;

export const SignupModal = (): JSX.Element => <LocationStateModal modalName="join" title="Sign up" padded={false}>
	<SignupForm />
</LocationStateModal>;

export const SignupPage = redirectIfLoggedIn("/", () => <Panel hoverLevel="low" style={{ margin: "auto", maxWidth: "30rem" }}>
	<PageTitle title="Sign Up" />
	<SignupForm />
</Panel>);
