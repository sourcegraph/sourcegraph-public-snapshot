// tslint:disable: typedef ordered-imports

import * as React from "react";
import Helmet from "react-helmet";
import {PlainRoute} from "react-router";

import {EventListener} from "sourcegraph/Component";
import {GlobalNav} from "sourcegraph/app/GlobalNav";
import "sourcegraph/components/styles/_normalize.css";
import * as styles from "sourcegraph/app/styles/App.css";

import {EventLogger, withEventLoggerContext, withViewEventsLogged} from "sourcegraph/util/EventLogger";
import {withFeaturesContext} from "sourcegraph/app/features";
import {withSiteConfigContext} from "sourcegraph/app/siteConfig";
import {desktopContainer} from "sourcegraph/desktop/DesktopContainer";

import {routes as homeRoutes} from "sourcegraph/home";
import {routes as pageRoutes} from "sourcegraph/page";
import {routes as styleguideRoutes} from "sourcegraph/styleguide";
import {routes as adminRoutes} from "sourcegraph/admin/routes";
import {routes as searchRoutes} from "sourcegraph/search/routes";
import {routes as userRoutes} from "sourcegraph/user";
import {routes as userSettingsRoutes} from "sourcegraph/user/settings/routes";
import {routes as repoRoutes} from "sourcegraph/repo/routes";
import {context} from "sourcegraph/app/context";

interface Props {
	main: JSX.Element;
	navContext: JSX.Element;
	location: any;
	params?: any;
	channelStatusCode?: number;
}

type State = any;

export class App extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	state = {
		className: "",
	};

	constructor(props: Props) {
		super(props);
		let className = styles.main_container;
		if (!context.user && location.pathname === "/") {
			className = styles.main_container_homepage;
		}
		this._handleSourcegraphDesktop = this._handleSourcegraphDesktop.bind(this);
		this.state = {
			className: className,
		};
	}

	_handleSourcegraphDesktop(event) {
		(this.context as any).router.replace(event.detail);
	}

	render(): JSX.Element | null {
		return (
			<div className={this.state.className}>
				<Helmet titleTemplate="%s · Sourcegraph" defaultTitle="Sourcegraph" />
				<GlobalNav desktop={desktopClient} params={this.props.params} location={this.props.location} channelStatusCode={this.props.channelStatusCode}/>
				{this.props.navContext && <div className={styles.breadcrumb}>{this.props.navContext}</div>}
				{this.props.main}
				<EventListener target={global.document} event="sourcegraph:desktop" callback={this._handleSourcegraphDesktop} />
			</div>
		);
	}
}

const desktopClient = global.document && navigator.userAgent.includes("Electron");
export const rootRoute: PlainRoute = {
	path: "/",
	component: withEventLoggerContext(EventLogger,
		withViewEventsLogged(
			withSiteConfigContext(
				withFeaturesContext(
					desktopClient ? desktopContainer(App) : App
				)
			)
		)
	),
	getIndexRoute: (location, callback) => {
		callback(null, homeRoutes);
	},
	getChildRoutes: (location, callback) => {
		callback(null, [
			...pageRoutes,
			...styleguideRoutes,
			...homeRoutes,
			...adminRoutes,
			...searchRoutes,
			...userRoutes,
			...userSettingsRoutes,
			...repoRoutes,
		]);
	},
};
