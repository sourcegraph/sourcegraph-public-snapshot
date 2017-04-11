import * as React from "react";

import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { colors } from "sourcegraph/components/utils";

import * as base from "sourcegraph/components/styles/_base.css";
import { Events, PAGE_LOCATION_GLOBAL_NAV } from "sourcegraph/tracking/constants/AnalyticsConstants";

export const SignupOrLogin = (props): JSX.Element => {

	const sx = Object.assign(
		{
			textAlign: "center",
			display: "inline-block",
			lineHeight: "calc(100% - 1px)",
			color: colors.blueGrayL1(),
		},
		props.style
	);

	return (
		<div style={sx}>
			<LocationStateToggleLink href="/login" modalName="login" location={location}
				onToggle={(v) => v && Events.LoginModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: PAGE_LOCATION_GLOBAL_NAV })}>
				Log in
			</LocationStateToggleLink>
			<span className={base.mh1}> or </span>
			<LocationStateToggleLink href="/join" modalName="join" location={location}
				onToggle={(v) => v && Events.JoinModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: PAGE_LOCATION_GLOBAL_NAV })}>
				Sign up
			</LocationStateToggleLink>
		</div>
	);
};
