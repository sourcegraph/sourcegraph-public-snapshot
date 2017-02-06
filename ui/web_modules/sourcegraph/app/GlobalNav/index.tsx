import { hover, keyframes } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import "sourcegraph/app/GlobalNav/GlobalNavBackend"; // for side-effects
import { SearchCTA, ShortcutCTA } from "sourcegraph/app/GlobalNav/GlobalNavCTA";
import { GlobalNavStore, SetQuickOpenVisible, SetShortcutMenuVisible } from "sourcegraph/app/GlobalNav/GlobalNavStore";
import { ShortcutModalComponent } from "sourcegraph/app/GlobalNav/ShortcutMenu";
import { SignupOrLogin } from "sourcegraph/app/GlobalNav/SignupOrLogin";
import { UserMenu } from "sourcegraph/app/GlobalNav/UserMenu";
import { AfterPrivateCodeSignup, BetaSignup, Login, Signup } from "sourcegraph/app/modals/index";
import { abs, isAtRoute } from "sourcegraph/app/routePatterns";
import { RouterContext, RouterLocation, getRepoFromRouter, getRevFromRouter } from "sourcegraph/app/router";
import { FlexContainer, Logo, TabItem, Tabs } from "sourcegraph/components";
import { colors, layout } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/index";
import { Container } from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import { IntegrationsContainer } from "sourcegraph/home/IntegrationsContainer";
import { DemoVideo } from "sourcegraph/home/modals/DemoVideo";
import { QuickOpenModal } from "sourcegraph/quickopen/Modal";
import { Store } from "sourcegraph/Store";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { isMobileUserAgent } from "sourcegraph/util/shouldPromptToInstallBrowserExtension";

interface Props {
	location: RouterLocation;
}

interface State {
	showSearch: boolean;
	showShortcut: boolean;
}

export class GlobalNav extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;

	constructor(props: Props) {
		super(props);
		this.onSearchDismiss = this.onSearchDismiss.bind(this);
		this.activateSearch = this.activateSearch.bind(this);
		this.state = { showSearch: false, showShortcut: false };
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.showSearch = GlobalNavStore.quickOpenVisible;
		state.showShortcut = GlobalNavStore.shortcutMenuVisible;
	}

	stores(): Store<any>[] {
		return [GlobalNavStore];
	}

	onSearchDismiss(): void {
		Dispatcher.Backends.dispatch(new SetQuickOpenVisible(false));
	}

	onShortcutDismiss(): void {
		Dispatcher.Backends.dispatch(new SetShortcutMenuVisible(false));
	}

	activateSearch(eventProps?: any): void {
		AnalyticsConstants.Events.Quickopen_Initiated.logEvent(eventProps);

		Dispatcher.Backends.dispatch(new SetQuickOpenVisible(true));
	}

	activateShortcutMenu(): void {
		AnalyticsConstants.Events.ShortcutMenu_Initiated.logEvent();
		Dispatcher.Backends.dispatch(new SetShortcutMenuVisible(true));
	}

	render(): JSX.Element {

		const hiddenNavRoutes = new Set([
			"/styleguide",
			"login",
			"join",
		]);

		const isHomeRoute = isAtRoute(this.context.router, abs.home);
		const shouldHide = hiddenNavRoutes.has(this.props.location.pathname) || (isHomeRoute && !context.user && context.authEnabled);

		const repo = getRepoFromRouter(this.context.router);
		const rev = getRevFromRouter(this.context.router);

		const sx = {
			backgroundColor: colors.white(),
			borderBottom: `1px solid ${colors.blueGray(0.3)}`,
			boxShadow: `${colors.blueGray(0.1)} 0px 1px 6px 0px`,
			display: shouldHide ? "none" : "block",
			zIndex: 100,
			paddingLeft: whitespace[2],
			paddingRight: whitespace[2],
		};

		const logoSpin = keyframes({
			"50%": { transform: "rotate(180deg) scale(1.2)" },
			"100%": { transform: "rotate(180deg) scale(1)" },
		});

		const modalName = (this.props.location.state && this.props.location.state["modal"]) || this.props.location.query["modal"];
		const modal = (
			<div>
				{modalName === "login" && !context.user &&
					<Login location={this.props.location} router={this.context.router} />}
				{modalName === "join" &&
					<Signup location={this.props.location} router={this.context.router} shouldHide={shouldHide} />}
				{modalName === "menuBeta" &&
					<BetaSignup location={this.props.location} router={this.context.router} />}
				{modalName === "afterPrivateCodeSignup" &&
					<AfterPrivateCodeSignup location={this.props.location} router={this.context.router} />}
				{modalName === "demo_video" &&
					<DemoVideo location={this.props.location} router={this.context.router} />}
				{modalName === "menuIntegrations" &&
					<IntegrationsContainer location={this.props.location} router={this.context.router} />}
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

				{/* TODO(john): the `|| null` is not very nice, we should avoid that. */}
				{isMobileUserAgent(navigator.userAgent) ? null : <ShortcutModalComponent onDismiss={this.onShortcutDismiss} showModal={this.state.showShortcut} activateShortcut={this.activateShortcutMenu} />}
				<QuickOpenModal repo={repo || null} rev={rev || null}
					showModal={this.state.showSearch}
					activateSearch={(eventProps) => this.activateSearch(eventProps)}
					onDismiss={this.onSearchDismiss} />
				<FlexContainer items="center" style={{ paddingRight: "0.5rem" }}>
					{isMobileUserAgent(navigator.userAgent) ? null : <a onClick={() => this.activateShortcutMenu()}><ShortcutCTA width={26} /></a>}
					<a onClick={() => this.activateSearch({ page_location: "SearchCTA" })}><SearchCTA width={18} /></a>
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
