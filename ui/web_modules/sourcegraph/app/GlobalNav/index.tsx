import { hover, keyframes } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import "sourcegraph/app/GlobalNav/GlobalNavBackend"; // for side-effects
import { GlobalNavStore, SetQuickOpenVisible } from "sourcegraph/app/GlobalNav/GlobalNavStore";
import { SearchCTA } from "sourcegraph/app/GlobalNav/SearchCTA";
import { SignupOrLogin } from "sourcegraph/app/GlobalNav/SignupOrLogin";
import { UserMenu } from "sourcegraph/app/GlobalNav/UserMenu";
import { BetaSignup, Login, Signup } from "sourcegraph/app/modals/index";
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

interface Props {
	location: RouterLocation;
}

interface State {
	showSearch: boolean;
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
		this.state = { showSearch: false };
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.showSearch = GlobalNavStore.quickOpenVisible;
	}

	stores(): Store<any>[] {
		return [GlobalNavStore];
	}

	onSearchDismiss(): void {
		Dispatcher.Backends.dispatch(new SetQuickOpenVisible(false));
	}

	activateSearch(eventProps?: any): void {
		AnalyticsConstants.Events.Quickopen_Initiated.logEvent(eventProps);

		Dispatcher.Backends.dispatch(new SetQuickOpenVisible(true));
	}

	render(): JSX.Element {

		const hiddenNavRoutes = new Set([
			"/",
			"/styleguide",
			"login",
			"join",
		]);

		const isHomeRoute = isAtRoute(this.context.router, abs.home);
		const dash = isHomeRoute && context.user;
		const shouldHide = hiddenNavRoutes.has(this.props.location.pathname) && !dash;

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

		let modal = <div />;
		if (this.props.location.state) {
			const m = this.props.location.state["modal"];
			modal = <div>
				{m === "login" && !context.user &&
					<Login location={this.props.location} router={this.context.router} />}
				{m === "join" &&
					<Signup location={this.props.location} router={this.context.router} shouldHide={shouldHide} />}
				{m === "menuBeta" &&
					<BetaSignup location={this.props.location} router={this.context.router} />}
				{m === "demo_video" &&
					<DemoVideo location={this.props.location} router={this.context.router} />}
				{m === "menuIntegrations" &&
					<IntegrationsContainer location={this.props.location} router={this.context.router} />}
			</div>;
		}
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
				<QuickOpenModal repo={repo || null} rev={rev || null}
					showModal={this.state.showSearch}
					activateSearch={(eventProps) => this.activateSearch(eventProps)}
					onDismiss={this.onSearchDismiss} />
				<FlexContainer items="center" style={{ paddingRight: "0.5rem" }}>
					<a onClick={() => this.activateSearch({ page_location: "SearchCTA" })}><SearchCTA width={14} /></a>
					{context.user
						? <UserMenu user={context.user} location={location} style={{ flex: "0 0 auto", marginTop: 4 }} />
						: <SignupOrLogin user={context.user} location={location} />
					}
				</FlexContainer>
			</FlexContainer>
		</div>;
	}
};
