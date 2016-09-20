import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";

/* TODO(chexee): abstract the presentational component from Modal */
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";

import {SignupForm} from "sourcegraph/user/Signup";

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
		<LocationStateModal modalName="join" location={props.location}
			onDismiss={(v) => EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "DismissJoinModal", {page_name: props.location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
			<div className={styles.modal} style={sx}>
				<SignupForm
					returnTo={props.shouldHide ? "/?ob=chrome" : props.location}
					location={props.location} />
			</div>
		</LocationStateModal>
	);
};
