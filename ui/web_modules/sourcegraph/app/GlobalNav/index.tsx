import { hover, keyframes } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { IDisposable } from "vs/base/common/lifecycle";

import { context } from "sourcegraph/app/context";
import "sourcegraph/app/GlobalNav/GlobalNavBackend"; // for side-effects
import { SearchCTA, ShortcutCTA } from "sourcegraph/app/GlobalNav/GlobalNavCTA";
import { GlobalNavStore } from "sourcegraph/app/GlobalNav/GlobalNavStore";
import { ShortcutModal } from "sourcegraph/app/GlobalNav/ShortcutMenu";
import { SignupOrLogin } from "sourcegraph/app/GlobalNav/SignupOrLogin";
import { UserMenu } from "sourcegraph/app/GlobalNav/UserMenu";
import { AfterSignup, BetaSignup, Login, Signup } from "sourcegraph/app/modals/index";
import { abs, isAtRoute } from "sourcegraph/app/routePatterns";
import { RouterContext, RouterLocation } from "sourcegraph/app/router";
import { FlexContainer, Logo, TabItem, Tabs } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { TourOverlay } from "sourcegraph/components/TourOverlay";
import { colors, layout } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/index";
import { Container } from "sourcegraph/Container";
import { toggleQuickopen } from "sourcegraph/editor/config";
import { IntegrationsContainer } from "sourcegraph/home/IntegrationsContainer";
import { Store } from "sourcegraph/Store";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { isMobileUserAgent } from "sourcegraph/util/shouldPromptToInstallBrowserExtension";
import { onWorkbenchShown } from "sourcegraph/workbench/main";

interface Props {
	location: RouterLocation;
}

interface State {
	showShortcut: boolean;
	workbenchShown: boolean;
}

export class GlobalNav extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;
	workbenchListener: IDisposable;

	constructor(props: Props) {
		super(props);
		this.activateSearch = this.activateSearch.bind(this);
		this.state = { showShortcut: false, workbenchShown: false };
	}

	componentDidMount(): void {
		super.componentDidMount();
		this.workbenchListener = onWorkbenchShown(shown => this.setState({ workbenchShown: shown } as State));
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();
		if (this.workbenchListener) {
			this.workbenchListener.dispose();
		}
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.showShortcut = GlobalNavStore.shortcutMenuVisible;
	}

	stores(): Store<any>[] {
		return [GlobalNavStore];
	}

	activateSearch(eventProps?: any): void {
		Events.Quickopen_Initiated.logEvent(eventProps);
		toggleQuickopen();
	}

	activateShortcutMenu(): void {
		Events.ShortcutMenu_Initiated.logEvent();
	}

	render(): JSX.Element {
		const hiddenNavRoutes = new Set([
			"/styleguide",
			"login",
			"join",
		]);

		const location = this.props.location;
		const router = this.context.router;

		const isHomeRoute = isAtRoute(router, abs.home);
		const shouldHide = hiddenNavRoutes.has(location.pathname) || (isHomeRoute && !context.user && context.authEnabled);

		const sx = {
			backgroundColor: colors.white(),
			borderBottom: `1px solid ${colors.blueGray(0.3)}`,
			boxShadow: `${colors.blueGray(0.1)} 0px 1px 6px 0px`,
			display: shouldHide ? "none" : "block",
			paddingLeft: whitespace[2],
			paddingRight: whitespace[2],
			zIndex: 2,
		};

		const logoSpin = keyframes({
			"50%": { transform: "rotate(180deg) scale(1.2)" },
			"100%": { transform: "rotate(180deg) scale(1)" },
		});

		const modalName = (location.state && location.state["modal"]) || location.query["modal"];
		const modal = (
			<div>
				{modalName === "login" && !context.user &&
					<Login />}
				{modalName === "join" &&
					<Signup />}
				{modalName === "menuBeta" &&
					<BetaSignup location={location} router={router} />}
				{modalName === "afterSignup" &&
					<AfterSignup />}
				{modalName === "menuIntegrations" &&
					<IntegrationsContainer location={location} router={router} />}
				<ShortcutModal location={location} router={router} />
			</div>
		);

		return <div
			{...layout.clearFix}
			id="global-nav"
			role="navigation"
			style={sx}>

			{modal}

			<FlexContainer justify="between" items="center">
				<FlexContainer items="center">
					<Link to="/" style={{ lineHeight: 0 }}>
						<div style={{ padding: whitespace[2], display: "inline-block" }}>
							<div {...hover({ animation: `${logoSpin} 0.5s ease-in-out 1` }) }>
								<Logo width="20px" />
							</div>
						</div>
					</Link>
					{context.user && <Tabs style={{ display: "inline-block", borderBottom: 0 }}>
						<Link to="/" style={{ outline: "none" }}>
							<TabItem active={isHomeRoute}>Repositories</TabItem>
						</Link>
					</Tabs>}
				</FlexContainer>
				{location.query["tour"] && <TourOverlay location={location} />}
				<FlexContainer items="center" style={{ paddingRight: "0.5rem" }}>
					{/* Only show the shortcut and search actions in the navbar when on a workbench view. */}
					{!isMobileUserAgent(navigator.userAgent) && this.state.workbenchShown &&
						<LocationStateToggleLink modalName="keyboardShortcuts" location={location} onToggle={this.activateShortcutMenu}>
							<ShortcutCTA width={26} />
						</LocationStateToggleLink>
					}
					{this.state.workbenchShown && <a onClick={() => this.activateSearch({ page_location: "SearchCTA" })}><SearchCTA width={18} /></a>}
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
