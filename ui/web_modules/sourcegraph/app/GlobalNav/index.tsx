import { hover, keyframes } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import { SearchCTA, ShortcutCTA } from "sourcegraph/app/GlobalNav/GlobalNavCTA";
import { SignupOrLogin } from "sourcegraph/app/GlobalNav/SignupOrLogin";
import { UserMenu } from "sourcegraph/app/GlobalNav/UserMenu";
import { RouterContext } from "sourcegraph/app/router";
import { FlexContainer, Logo } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { colors, layout } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/index";
import { toggleQuickopen } from "sourcegraph/editor/config";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { isMobileUserAgent } from "sourcegraph/util/shouldPromptToInstallBrowserExtension";

interface State {
	showShortcut: boolean;
	workbenchShown: boolean;
}

export const hiddenNavRoutes = new Set([
	"/styleguide",
	"/enterprise",
	"login",
	"join",
]);

export class GlobalNav extends React.Component<{}, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;
	state: State = { showShortcut: false, workbenchShown: false };

	activateSearch = (): void => {
		Events.Quickopen_Initiated.logEvent({ page_location: "SearchCTA" });
		toggleQuickopen();
	}

	activateShortcutMenu = (): void => {
		Events.ShortcutMenu_Initiated.logEvent();
	}

	render(): JSX.Element {
		const router = this.context.router;
		const location = router.location;

		const sx = {
			backgroundColor: colors.white(),
			borderBottom: `1px solid ${colors.blueGray(0.3)}`,
			boxShadow: `${colors.blueGray(0.1)} 0px 1px 6px 0px`,
			paddingLeft: whitespace[2],
			paddingRight: whitespace[2],
			zIndex: 2000,
		};

		const logoSpin = keyframes({
			"50%": { transform: "rotate(180deg) scale(1.2)" },
			"100%": { transform: "rotate(180deg) scale(1)" },
		});
		return <div
			{...layout.clearFix}
			id="global-nav"
			role="navigation"
			style={sx}>

			<FlexContainer justify="between" items="center">
				<FlexContainer items="center">
					<Link to="/" style={{ lineHeight: 0 }}>
						<div style={{ padding: whitespace[2], display: "inline-block" }}>
							<div {...hover({ animation: `${logoSpin} 0.5s ease-in-out 1` }) }>
								<Logo width="20px" />
							</div>
						</div>
					</Link>
				</FlexContainer>
				<FlexContainer items="center" style={{ paddingRight: "0.5rem" }}>
					{/* Only show the shortcut and search actions in the navbar when on a workbench view. */}
					{!isMobileUserAgent(navigator.userAgent) &&
						<LocationStateToggleLink modalName="keyboardShortcuts" location={location} onToggle={this.activateShortcutMenu}>
							<ShortcutCTA width={26} />
						</LocationStateToggleLink>
					}
					{this.state.workbenchShown && <a onClick={this.activateSearch}><SearchCTA width={18} /></a>}
					{context.authEnabled &&
						(context.user
							? <UserMenu user={context.user} location={location} style={{ flex: "0 0 auto", marginTop: 4 }} />
							: <SignupOrLogin user={context.user} location={location} />)
					}
				</FlexContainer>
			</FlexContainer>
		</div>;
	}
};
