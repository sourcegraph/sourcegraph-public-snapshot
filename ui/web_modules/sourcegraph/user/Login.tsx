import * as React from "react";
import { Panel, SignupLoginAuth } from "sourcegraph/components";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { redirectIfLoggedIn } from "sourcegraph/user/redirectIfLoggedIn";

const Login = () => <SignupLoginAuth>
	To log in, authorize with GitHub:
</SignupLoginAuth>;

export const LoginModal = (): JSX.Element => <LocationStateModal modalName="login" title="Log in" padded={false}>
	<Login />
</LocationStateModal>;

export const LoginPage = redirectIfLoggedIn("/", () => <Panel hoverLevel="low" style={{ margin: "auto", maxWidth: "30rem" }}>
	<PageTitle title="Sign In" />
	<Login />
</Panel>);
