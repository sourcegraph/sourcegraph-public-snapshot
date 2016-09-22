import {Location} from "history";
import * as React from "react";
import {Link} from "react-router";
import {InjectedRouter} from "react-router";

import {Base, FlexContainer, Logo} from "sourcegraph/components";
import {CheckIcon, EllipsisHorizontal} from "sourcegraph/components/Icons";
import {colors, layout} from "sourcegraph/components/utils";

import * as styles from "sourcegraph/app/styles/GlobalNav.css";

import {BetaSignup, Integrations, Login, Signup} from "sourcegraph/app/modals/index";
import {DemoVideo} from "sourcegraph/home/modals/DemoVideo";

import {SearchForm} from "sourcegraph/app/GlobalNav/SearchForm";
import {SignupOrLogin} from "sourcegraph/app/GlobalNav/SignupOrLogin";
import {UserMenu} from "sourcegraph/app/GlobalNav/UserMenu";

import {context} from "sourcegraph/app/context";
import {LocationState} from "sourcegraph/app/locationState";
import {isRootRoute, rel} from "sourcegraph/app/routePatterns";
import {repoParam, repoPath} from "sourcegraph/repo";

const hiddenNavRoutes = new Set([
	"/",
	"/styleguide",
	"login",
	"join",
]);

interface Props {
	navContext?: JSX.Element;
	location: Location;
	params: any;
	channelStatusCode?: number;
	role?: string;
	desktop: boolean;
}

interface Context {
	router: InjectedRouter;
}

export function GlobalNav(
	{desktop, navContext, location, params, channelStatusCode}: Props,
	{router}: Context
): JSX.Element {

	const dash = location.pathname.match(/^\/?$/) && context.user;
	const shouldHide = hiddenNavRoutes.has(location.pathname) && !dash;
	const showSearchForm = !dash || desktop;

	const repoRev = repoParam(params.splat);
	const repo = repoRev ? repoPath(repoRev) : null;

	const sx = {
		backgroundColor: colors.white(),
		borderBottom: `1px solid ${colors.coolGray3(0.3)}`,
		boxShadow: `${colors.coolGray3(0.1)} 0px 1px 6px 0px`,
		display: shouldHide ? "none" : "block",
		zIndex: 100,
	};

	return (
		<Base
			{...layout.clearFix}
			id="global-nav"
			role="navigation"
			px={2}
			py={1}
			style={sx}>

			{location.state &&
				<div>

					{(location.state as LocationState).modal === "login" && !context.user &&
						<Login location={location}/> }

					{(location.state as LocationState).modal === "join" &&
						<Signup location={location} router={router} shouldHide={shouldHide}/> }

					{(location.state as LocationState).modal === "menuBeta" &&
						<BetaSignup location={location} router={router} />}

					{(location.state as LocationState).modal === "menuIntegrations" &&
						<Integrations location={location} router={router} />}

					{(location.state as LocationState).modal === "demo_video" && isRootRoute(location) &&
						// TODO(mate, chexee): consider moving this to Home.tsx
						<DemoVideo location={location} />}

				</div>
			}

			<FlexContainer justify="between" items="center">
				<Link to="/" style={{lineHeight: "0"}}>
					<Base p={2}>
						<Logo className={styles.logomark}
						width="20px"
						type="logomark"/>
					</Base>
				</Link>

				{typeof channelStatusCode !== "undefined" && channelStatusCode === 0 && <EllipsisHorizontal className={styles.icon_ellipsis} title="Your editor could not identify the symbol"/>}
				{typeof channelStatusCode !== "undefined" && channelStatusCode === 1 && <CheckIcon className={styles.icon_check} title="Sourcegraph successfully looked up symbol" />}

				{showSearchForm && <SearchForm repo={repo} location={location} router={router} showResultsPanel={location.pathname !== `/${rel.search}`} style={{flex: "1 1 100%", margin: "0 8px"}} />}

				{context.user
					? <UserMenu user={context.user} location={location} />
					: <SignupOrLogin user={context.user} location={location} />
				}
			</FlexContainer>
		</Base>
	);
}

(GlobalNav as React.StatelessComponent<Props>).contextTypes = {
	router: React.PropTypes.object.isRequired,
};
