import * as React from "react";
import { Link } from "react-router";
import { Avatar, FlexContainer, Heading, Menu, Popover } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { ChevronDown } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { Events, PAGE_LOCATION_GLOBAL_NAV } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";

export function handleIntercomToggle(locationOnPage: string): void {
	global.window.Intercom("show");
	Events.ContactIntercom_Clicked.logEvent({
		page_name: location.pathname,
		location_on_page: locationOnPage,
	}
	);
}

// Events being logged on this page
const logBetaModal = () => Events.BetaModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: PAGE_LOCATION_GLOBAL_NAV });
const logSignout = () => Events.Logout_Clicked.logEvent(); EventLogger.logout();
const logBrowserExt = () => Events.ToolsModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: PAGE_LOCATION_GLOBAL_NAV });

const footLinkSx = {
	...typography.small,
	color: colors.blueGray(),
	paddingRight: whitespace[3],
};

export const UserMenu = (props): JSX.Element => {
	return (
		<div style={{ display: "inline-block", padding: whitespace[2] }}>
			<Popover left={true}>
				{ /* CAUTION: If you change the sourcegraph-user-menu class, you may break an e2e test. */}
				<FlexContainer className="sourcegraph-user-menu" items="center" style={{ lineHeight: "0", height: 29 }}>
					{props.user.AvatarURL ? <Avatar size="small" img={props.user.AvatarURL} /> : <div>{props.user.Login}</div>}
					<ChevronDown color={colors.blueGray()} style={{ marginLeft: 8 }} />
				</FlexContainer>
				<Menu style={{ padding: 0, position: "relative", zIndex: 2, width: 220 }}>
					<Heading level={7} color="gray" style={{
						padding: 0,
						marginBottom: whitespace[2],
						marginTop: whitespace[3],
					}}>Signed in as</Heading>
					<div style={{ marginBottom: whitespace[3] }}>{props.user.Login}</div>
					<hr role="divider" />
					<Link to="/settings" role="menu_item">Accounts and billing</Link>
					<LocationStateToggleLink
						modalName="menuIntegrations"
						role="menu_item"
						location={location}
						onToggle={logBrowserExt}>
						Browser extensions
					</LocationStateToggleLink>
					<LocationStateToggleLink
						href="/beta"
						modalName="menuBeta"
						role="menu_item"
						location={location}
						onToggle={logBetaModal}>
						Beta program
					</LocationStateToggleLink>
					<a href="/docs" role="menu_item">Docs</a>
					<a href="/pricing" role="menu_item">Pricing</a>
					<a onClick={() => handleIntercomToggle(PAGE_LOCATION_GLOBAL_NAV)} role="menu_item">
						Contact
					</a>
					<hr role="divider" />
					<a role="menu_item" href="/-/logout" onClick={logSignout}>Sign out</a>
					<hr role="divider" />
					<div style={{ paddingTop: whitespace[1], paddingBottom: whitespace[3], textAlign: "center" }}>
						<Link to="/security" style={footLinkSx}>Security</Link>
						<Link to="/privacy" style={footLinkSx}>Privacy</Link>
						<Link to="/terms" style={{ ...footLinkSx, padding: 0 }}>Terms</Link>
					</div>
				</Menu>
			</Popover>
		</div>
	);
};
