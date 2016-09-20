import * as React from "react";

import {EventLogger} from "sourcegraph/util/EventLogger";

import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";

import * as base from "sourcegraph/components/styles/_base.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export const SignupOrLogin = (props): JSX.Element => {

	const sx = {
		flex: "1 0 135px",
		textAlign: "center",
	};

	return(
		<div style={sx}>
			<LocationStateToggleLink href="/login" modalName="login" location={location}
				onToggle={(v) => v && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ShowLoginModal", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
				Log in
			</LocationStateToggleLink>
			<span className={base.mh1}> or </span>
			<LocationStateToggleLink href="/join" modalName="join" location={location}
				onToggle={(v) => v && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ShowSignUpModal", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
				Sign up
			</LocationStateToggleLink>
		</div>
	);
};
