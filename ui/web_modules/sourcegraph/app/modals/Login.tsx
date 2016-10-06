import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";
import {LocationStateModal} from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import {LoginForm} from "sourcegraph/user/Login";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	location: Location;
	router: InjectedRouter;
}

export const Login = (props: Props): JSX.Element => {
	const sx = {
		maxWidth: "380px",
		marginLeft: "auto",
		marginRight: "auto",
	};

	return(
		<LocationStateModal modalName="login" location={props.location} router={props.router}
			onDismiss={(v) => EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "DismissLoginModal", {page_name: props.location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
			<div className={styles.modal} style={sx}>
				<LoginForm
					returnTo={props.location}
					location={props.location} />
			</div>
		</LocationStateModal>
	);
};
