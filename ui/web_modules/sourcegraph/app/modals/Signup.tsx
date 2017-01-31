import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { Heading } from "sourcegraph/components";
import { LocationStateModal } from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import { SignupForm, defaultOnboardingPath } from "sourcegraph/user/Signup";

interface Props {
	location: RouterLocation;
	router: Router;
	shouldHide: boolean;
}

export const Signup = (props: Props): JSX.Element => {
	const sx = {
		maxWidth: "420px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	let newUserPath = props.location.pathname.indexOf("/-/blob/") !== -1 ? { pathname: props.location.pathname, hash: props.location.hash } : defaultOnboardingPath;
	return (
		<LocationStateModal modalName="join" location={props.location} router={props.router}>
			<div className={styles.modal} style={sx}>
				<Heading level={3} align="center" underline="orange">Get started with Sourcegraph</Heading>
				<SignupForm
					newUserReturnTo={newUserPath}
					returnTo={props.shouldHide ? "/" : props.location.pathname}
					location={props.location} />
			</div>
		</LocationStateModal>
	);
};
