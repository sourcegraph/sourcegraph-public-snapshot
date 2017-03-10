import * as React from "react";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { LoginForm } from "sourcegraph/user/Login";

export const Login = (): JSX.Element => <LocationStateModal modalName="login" title="Log in" padded={false}>
	<LoginForm />
</LocationStateModal>;
