import * as React from "react";
import {InjectedRouter} from "react-router";

import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import {Location} from "sourcegraph/Location";
import {SignupForm} from "sourcegraph/user/Signup";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	location: Location;
	router: InjectedRouter;
	shouldHide: boolean;
}

export const Signup = (props: Props): JSX.Element => {
	const sx = {
		maxWidth: "380px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return(
		<LocationStateModal modalName="join" location={props.location} router={props.router}
			onDismiss={(v) => EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "DismissJoinModal", {page_name: props.location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
			<div className={styles.modal} style={sx}>
				<SignupForm
					returnTo={props.shouldHide ? "/" : props.location.pathname}
					queryObj={props.shouldHide ? {ob: "chrome"} : Object.assign({}, props.location.query)}
					location={props.location} />
			</div>
		</LocationStateModal>
	);
};
