import "sourcegraph/components/styles/_normalize.css";

import * as React from "react";
import * as ReactDOM from "react-dom";
import { PlainRoute } from "react-router";

import { context } from "sourcegraph/app/context";
import { hiddenNavRoutes } from "sourcegraph/app/GlobalNav";
import { abs, isAtRoute, rel } from "sourcegraph/app/routePatterns";
import { Router, RouterLocation, getRepoFromRouter, setRouter } from "sourcegraph/app/router";
import { pageRoutes } from "sourcegraph/app/routes/pageRoutes";
import { repoRoutes } from "sourcegraph/app/routes/repoRoutes";
import { styleguideRoutes } from "sourcegraph/app/routes/styleguideRoutes";
import { symbolRoutes } from "sourcegraph/app/routes/symbolRoutes";
import { userRoutes } from "sourcegraph/app/routes/userRoutes";
import { userSettingsRoutes } from "sourcegraph/app/routes/userSettingsRoutes";
import * as styles from "sourcegraph/app/styles/App.css";
import { HomeRouter } from "sourcegraph/home/HomeRouter";
import { withViewEventsLogged } from "sourcegraph/tracking/withViewEventsLogged";
import { updateConfiguration } from "sourcegraph/workbench/ConfigurationService";
import { Services } from "sourcegraph/workbench/services";
import { RouterContext } from "sourcegraph/workbench/utils";
import { Workbench } from "sourcegraph/workbench/workbench";

import { $ } from "vs/base/browser/builder";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { DynamicOverlay } from "vs/workbench/parts/content/overlay/dynamicOverlay";

import { ShortcutModal } from "sourcegraph/app/GlobalNav/ShortcutMenu";
import { AfterSignup, BetaSignup } from "sourcegraph/app/modals/index";
import { IntegrationsContainer } from "sourcegraph/home/IntegrationsContainer";
import { LoginModal } from "sourcegraph/user/Login";
import { SignupModal } from "sourcegraph/user/Signup";

interface Props {
	main: JSX.Element;
	injectedComponent: JSX.Element;
	footer?: JSX.Element;
	location?: RouterLocation;
}

export class App extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };
	overlay: DynamicOverlay | undefined;
	renderedElement: JSX.Element | undefined;
	renderedContainer: HTMLElement;

	constructor(props: Props, context: { router: Router }) {
		super(props, context);
		setRouter(context.router);
		this.setupWorkbenchContainer();
	}

	componentWillReceiveProps(newProps: Props): void {
		// If navigating to a repo page, ensure injected overlays are destroyed.
		if (getRepoFromRouter(this.context.router)) {
			this.disposeInjectedComponent();
		}
		// If the new component is different than the previous set rendered element to undefined
		// so we create a new element in the component update loop.
		if (this.renderedElement && this.props.injectedComponent && newProps.injectedComponent && newProps.injectedComponent.type !== this.props.injectedComponent.type ||
			!newProps.location || !this.props.location || newProps.location.pathname !== this.props.location.pathname) {
			this.renderedElement = undefined;
		}
	}

	disposeInjectedComponent(): void {
		if (this.overlay) {
			ReactDOM.unmountComponentAtNode(this.renderedContainer);
			this.overlay.destroy();
			this.overlay = undefined;
		}
	}

	/**
	 * setupWorkbenchContainer is responsible for configuring the container passed to the workbench dynamic content constructor
	 */
	setupWorkbenchContainer(): void {
		this.renderedContainer = document.createElement("div");
		this.renderedContainer.style.display = "flex";
		this.renderedContainer.style.flexDirection = "column";
		this.renderedContainer.style.flex = "1";
		this.renderedContainer.style.backgroundColor = "rgba(242, 244, 248, 0.2)";
		this.renderedContainer.style.fontSize = "16px";
		this.renderedContainer.style.color = "rgba(35, 48, 67, 1)";
		this.renderedContainer.style.lineHeight = "1.5";
		this.renderedContainer.style.overflowY = "scroll";
	}

	renderInjectedComponent(injectedComponent: JSX.Element): void {
		updateConfiguration((config: any) => {
			config.workbench.statusBar.visible = false;
		});
		this.disposeInjectedComponent();
		const instantiationService = Services.get(IInstantiationService) as IInstantiationService;
		this.overlay = instantiationService.createInstance(DynamicOverlay);
		ReactDOM.render(injectedComponent, this.renderedContainer);

		// Adjust render for if the nav bar is going to be visible.
		const isHomeRoute = isAtRoute(this.context.router, abs.home);
		const shouldHide = hiddenNavRoutes.has(location.pathname) || (isHomeRoute && !context.user && context.authEnabled);
		const height = shouldHide ? 0 : 45;

		this.overlay.create($(this.renderedContainer), Object.assign({}, this.overlay.getDefaultOverlayStyles(), { height: `calc(100% - ${height}px)`, top: 0 }));
		this.overlay.show("flex");
	}

	render(): JSX.Element {
		let className = styles.main_container;
		if (!context.user && this.context.router.location.pathname === "/") {
			className = styles.main_container_homepage;
		}
		if (!this.renderedElement) {
			this.renderedElement = this.props.main;
			if (this.props.injectedComponent) {
				this.renderedElement = React.cloneElement(this.props.main, { injectedComponentCallback: this.renderInjectedComponent.bind(this, <RouterContext><div style={{ height: "100%", display: "flex", flexDirection: "column" }}>{this.props.injectedComponent}{this.props.footer}</div></RouterContext>) });
			}
		}
		const location = this.context.router.location;
		const modalName = (location.state && location.state["modal"]) || location.query["modal"];
		const router = this.context.router;
		const modal = (
			<div>
				{modalName === "login" && !context.user &&
					<LoginModal />}
				{modalName === "join" &&
					<SignupModal />}
				{modalName === "menuBeta" &&
					<BetaSignup location={location} router={router} />}
				{modalName === "afterSignup" &&
					<AfterSignup />}
				{modalName === "menuIntegrations" &&
					<IntegrationsContainer location={location} router={router} />}
				<ShortcutModal location={location} router={router} />
			</div>
		);

		return (
			<div className={className}>
				{modal}
				{this.renderedElement}
			</div>
		);
	}
}

export const rootRoute: PlainRoute = {
	path: "/",
	component: withViewEventsLogged(App),
	getIndexRoute: (location, callback) => {
		callback(null, {
			path: rel.home,
			getComponents: (loc, cb) => {
				cb(null, { main: Workbench, injectedComponent: HomeRouter });
			},
		});
	},
	getChildRoutes: (location, callback) => {
		callback(null, [
			...pageRoutes,
			...styleguideRoutes,
			...userRoutes,
			...userSettingsRoutes,
			...symbolRoutes,
			...repoRoutes,
		]);
	},
};
