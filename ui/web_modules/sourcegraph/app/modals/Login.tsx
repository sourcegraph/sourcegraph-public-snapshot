import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { LoginForm } from "sourcegraph/user/Login";

interface Props {
	location: RouterLocation;
}

export const Login = (props: Props): JSX.Element => {

	return (
		<LocationStateModal modalName="login" title="Log in">
			<LoginForm
				returnTo={props.location}
				location={props.location} />
		</LocationStateModal>
	);
};
