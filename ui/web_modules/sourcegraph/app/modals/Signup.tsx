import * as React from "react";
import {InjectedRouter} from "react-router";

import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import {Location} from "sourcegraph/Location";
import {SignupForm} from "sourcegraph/user/Signup";

interface Props {
	location: Location;
	router: InjectedRouter;
	shouldHide: boolean;
}

const defaultOnboardingPath = "/github.com/sourcegraph/checkup/-/blob/checkup.go#L153";

export const Signup = (props: Props): JSX.Element => {
	const sx = {
		maxWidth: "380px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	let newUserPath = props.location.pathname.indexOf("/-/blob/") !== -1 ? props.location.pathname : defaultOnboardingPath;
	return(
		<LocationStateModal modalName="join" location={props.location} router={props.router}>
			<div className={styles.modal} style={sx}>
				<SignupForm
					newUserReturnTo={newUserPath}
					returnTo={props.shouldHide ? "/" : props.location.pathname}
					queryObj={props.shouldHide ? {ob: "chrome"} : Object.assign({}, props.location.query)}
					location={props.location} />
			</div>
		</LocationStateModal>
	);
};
