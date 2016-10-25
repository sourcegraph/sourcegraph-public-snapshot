import * as classNames from "classnames";
import * as React from "react";
import {Link} from "react-router";
import {Avatar, Base, FlexContainer, Heading, Menu, Popover} from "sourcegraph/components";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import * as base from "sourcegraph/components/styles/_base.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import {ChevronDown} from "sourcegraph/components/symbols/Zondicons";
import {colors} from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

export const UserMenu = (props): JSX.Element => {
	//@TODO: Check if this works in staging
	function handleIntercomToggle(): void {
		global.window.Intercom("show");
		EventLogger.logEventForCategory(
			AnalyticsConstants.CATEGORY_AUTH,
			AnalyticsConstants.ACTION_CLICK, "ClickContactIntercom", {
				page_name: location.pathname,
				location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV,
			}
		);
	}

	return(
		<Base p={2} style={{display: "inline-block"}}>
			<Popover left={true}>
				<FlexContainer items="center" style={{lineHeight: "0", height: 29}}>
					{props.user.AvatarURL ? <Avatar size="small" img={props.user.AvatarURL} /> : <div>{props.user.Login}</div>}
					<ChevronDown width={12} color={colors.coolGray3()} style={{marginLeft: "8px"}}/>
				</FlexContainer>
				<Menu className={classNames(base.pa0, base.mr2)} style={{width: "220px"}}>
					<div className={classNames(base.pa0, base.mb2, base.mt3)}>
						<Heading level={7} color="gray">Signed in as</Heading>
					</div>
					<div>{props.user.Login}</div>
					<hr role="divider" className={base.mv3} />
					<Link to="/settings" role="menu_item">Settings</Link>
					<LocationStateToggleLink href="/integrations" modalName="menuIntegrations" role="menu_item" location={location}	onToggle={(v) => v && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ClickToolsandIntegrations", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
						Tools and integrations
					</LocationStateToggleLink>
					<LocationStateToggleLink href="/beta" modalName="menuBeta" role="menu_item" location={location}	onToggle={(v) => v && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ClickJoinBeta", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
						Beta program
					</LocationStateToggleLink>
					<a href="/docs" role="menu_item" onClick={(v) => v && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "ClickJoinBeta", {page_name: location.pathname, location_on_page: AnalyticsConstants.PAGE_LOCATION_GLOBAL_NAV})}>
						Docs
					</a>
					<a onClick={handleIntercomToggle} role="menu_item">
						Contact
					</a>
					<hr role="divider" className={base.mt3} />
					<a role="menu_item" href="/-/logout" onClick={(e) => { EventLogger.logout(); }}>Sign out</a>
					<hr role="divider" className={base.mt2} />
					<div className={classNames(base.pv1, base.mb1, typography.tc)}>
						<Link to="/security" className={classNames(typography.f7, typography.link_subtle, base.pr3)}>Security</Link>
						<Link to="/privacy" className={classNames(typography.f7, typography.link_subtle, base.pr3)}>Privacy</Link>
						<Link to="/terms" className={classNames(typography.f7, typography.link_subtle)}>Terms</Link>
					</div>
				</Menu>
			</Popover>
		</Base>
	);
};
