import { hover, merge } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { RouterLocation } from "sourcegraph/app/router";
import { Button, FlexContainer, Logo } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { Events, PAGE_LOCATION_GLOBAL_NAV } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface Props {
	context: any;
	location: RouterLocation;
	style?: React.CSSProperties;
}

const navItemSx = {
	marginTop: whitespace[2],
	marginRight: whitespace[5],
	color: colors.blueGray(),
};

const navHover = hover({ color: colors.blueGrayD1() });

export function Nav({ context, style, location }: Props): JSX.Element {

	return <FlexContainer justify="between" wrap={true} style={style}>
		<Logo type="logotype" width="195" />

		<div>
			<Link to="/about" {...merge(navItemSx, navHover, { marginLeft: 0 }) }>About</Link>
			<Link to="/pricing" {...merge(navItemSx, navHover) }>Pricing</Link>
			<a href="/jobs" {...merge(navItemSx, navHover) } onClick={() => Events.JobsCTA_Clicked.logEvent()}>Jobs</a>

			{!(context as any).signedIn &&
				<LocationStateToggleLink
					href="/login"
					modalName="login"
					location={location}
					onToggle={(v) => v && Events.LoginModal_Initiated.logEvent({ page_name: location.pathname })}
					{...merge(navItemSx, navHover) }
				>Log in</LocationStateToggleLink>
			}

			{!(context as any).signedIn &&
				<LocationStateToggleLink
					href="/join"
					modalName="join"
					location={location}
					onToggle={(v) => v && Events.JoinModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: PAGE_LOCATION_GLOBAL_NAV })}
					{ ...layout.hide.sm }
					style={{
						paddingTop: whitespace[2],
						paddingBottom: whitespace[2],
					}}
				><Button color="orange" size="small" style={{ marginTop: whitespace[1] }}>Sign up</Button>
				</LocationStateToggleLink>
			}
		</div>
	</FlexContainer>;
};
