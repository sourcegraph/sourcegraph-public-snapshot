import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { LocationStateModal } from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import { LoginForm } from "sourcegraph/user/Login";

interface Props {
	location: RouterLocation;
	router: Router;
}

export const Login = (props: Props): JSX.Element => {
	const sx = {
		maxWidth: "380px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return (
		<LocationStateModal modalName="login" location={props.location} router={props.router}>
			<div className={styles.modal} style={sx}>
				<LoginForm
					returnTo={props.location}
					location={props.location} />
			</div>
		</LocationStateModal>
	);
};
