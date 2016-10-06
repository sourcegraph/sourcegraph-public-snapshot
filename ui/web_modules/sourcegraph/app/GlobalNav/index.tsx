import {Location} from "history";
import * as React from "react";
import {Link} from "react-router";
import {InjectedRouter} from "react-router";
import {context} from "sourcegraph/app/context";
import  "sourcegraph/app/GlobalNav/GlobalNavBackend"; // for side-effects
import {GlobalNavStore, SetQuickOpenVisible} from "sourcegraph/app/GlobalNav/GlobalNavStore";
import {SearchCTA} from "sourcegraph/app/GlobalNav/SearchCTA";
import {SignupOrLogin} from "sourcegraph/app/GlobalNav/SignupOrLogin";
import {UserMenu} from "sourcegraph/app/GlobalNav/UserMenu";
import {LocationState} from "sourcegraph/app/locationState";
import {BetaSignup, Integrations, Login, Signup} from "sourcegraph/app/modals/index";
import {isRootRoute} from "sourcegraph/app/routePatterns";
import * as styles from "sourcegraph/app/styles/GlobalNav.css";
import {Base, FlexContainer, Logo} from "sourcegraph/components";
import {colors, layout} from "sourcegraph/components/utils";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {DemoVideo} from "sourcegraph/home/modals/DemoVideo";
import {QuickOpenModal} from "sourcegraph/quickopen/Modal";
import {repoParam, repoPath, repoRev} from "sourcegraph/repo";
import {Store} from "sourcegraph/Store";

interface Props {
	navContext?: JSX.Element;
	location: Location & {state: LocationState};
	params: any;
	role?: string;
}

interface State extends Props {
	showSearch: boolean;
}

export class GlobalNav extends Container<Props, State> {

	constructor(props: Props) {
		super(props);
		this.onSearchDismiss = this.onSearchDismiss.bind(this);
		this.activateSearch = this.activateSearch.bind(this);
		this.state = Object.assign({}, props, {
			showSearch: false,
		});
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.showSearch = GlobalNavStore.quickOpenVisible;
	}

	stores(): Store<any>[] {
		return [GlobalNavStore];
	}

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

	onSearchDismiss(): void {
		Dispatcher.Backends.dispatch(new SetQuickOpenVisible(false));
	}

	activateSearch(): void {
		Dispatcher.Backends.dispatch(new SetQuickOpenVisible(true));
	}

	render(): JSX.Element {
		const {location, params} = this.props;

		const hiddenNavRoutes = new Set([
			"/",
			"/styleguide",
			"login",
			"join",
		]);

		const dash = isRootRoute(location) && context.user;
		const shouldHide = hiddenNavRoutes.has(location.pathname) && !dash;

		const revSpec = repoParam(params.splat);
		const [repo, rev] = revSpec ?
			[repoPath(revSpec), repoRev(revSpec)] :
			[null, null];

		const sx = {
			backgroundColor: colors.white(),
			borderBottom: `1px solid ${colors.coolGray3(0.3)}`,
			boxShadow: `${colors.coolGray3(0.1)} 0px 1px 6px 0px`,
			display: shouldHide ? "none" : "block",
			zIndex: 100,
		};

		let modal = <div />;
		if (location.state) {
			const m = location.state.modal;
			modal = <div>
				{m === "login" && !context.user && <Login location={location} router={this.context.router} />}
				{m === "join" && <Signup location={location} router={this.context.router} shouldHide={shouldHide} />}
				{m === "menuBeta" && <BetaSignup location={location} router={this.context.router} />}
				{m === "menuIntegrations" && <Integrations location={location} router={this.context.router} />}
				{m === "demo_video" && <DemoVideo location={location} router={this.context.router} />}
			</div>;
		}
		return <Base
			{...layout.clearFix}
			id="global-nav"
			role="navigation"
			px={2}
			py={1}
			style={sx}>

			{modal}

			<FlexContainer justify="between" items="center">
				<Link to="/" style={{lineHeight: "0"}}>
					<Base p={2}>
						<Logo className={styles.logomark}
						width="20px"
						type="logomark"/>
					</Base>
				</Link>

				<QuickOpenModal repo={repo} rev={rev}
					showModal={this.state.showSearch}
					activateSearch={this.activateSearch}
					onDismiss={this.onSearchDismiss} />
				<FlexContainer items="center" style={{paddingRight: "0.5rem"}}>
					{location.pathname !== "/" && <a onClick={this.activateSearch}><SearchCTA width={14} /></a>}
					{context.user
						? <UserMenu user={context.user} location={location} style={{flex: "0 0 auto", marginTop: 4}} />
						: <SignupOrLogin user={context.user} location={location} />
					}
				</FlexContainer>
			</FlexContainer>
		</Base>;
	}
};
