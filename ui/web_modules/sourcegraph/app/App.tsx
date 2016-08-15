// tslint:disable: typedef ordered-imports

import * as React from "react";
import Helmet from "react-helmet";

import {EventListener} from "sourcegraph/Component";
import {GlobalNav} from "sourcegraph/app/GlobalNav";
import {Footer} from "sourcegraph/app/Footer";
import "sourcegraph/components/styles/_normalize.css";
import * as styles from "sourcegraph/app/styles/App.css";

import {EventLogger, withEventLoggerContext, withViewEventsLogged} from "sourcegraph/util/EventLogger";
import {withFeaturesContext} from "sourcegraph/app/features";
import {withSiteConfigContext} from "sourcegraph/app/siteConfig";
import {withUserContext} from "sourcegraph/app/user";
import {withAppdashRouteStateRecording} from "sourcegraph/app/appdash";
import {withChannelListener} from "sourcegraph/channel/withChannelListener";
import {redirectForDashboard} from "sourcegraph/app/dashboardRedirect";

import {routes as homeRoutes} from "sourcegraph/home/index";
import {routes as dashboardRoutes} from "sourcegraph/dashboard/index";
import {routes as pageRoutes} from "sourcegraph/page/index";
import {routes as styleguideRoutes} from "sourcegraph/styleguide/index";
import {routes as desktopRoutes} from "sourcegraph/desktop/index";
import {routes as channelRoutes} from "sourcegraph/channel/index";
import {route as miscRoute} from "sourcegraph/misc/golang";
import {routes as adminRoutes} from "sourcegraph/admin/routes";
import {routes as searchRoutes} from "sourcegraph/search/routes";
import {routes as userRoutes} from "sourcegraph/user/index";
import {routes as userSettingsRoutes} from "sourcegraph/user/settings/routes";
import {routes as repoRoutes} from "sourcegraph/repo/routes";

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
		signedIn: React.PropTypes.bool.isRequired,
	};

	state = {
		className: "",
	};

	constructor(props: Props, context) {
		super(props);
		let className = styles.main_container;
		if (!context.signedIn && location.pathname === "/") {
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
				<GlobalNav params={this.props.params} location={this.props.location} channelStatusCode={this.props.channelStatusCode}/>
				<div className={styles.main_content} id="scroller" ref="mainContent">
					<div className={styles.inner_main_content}>
						{this.props.navContext && <div className={styles.breadcrumb}>{this.props.navContext}</div>}
						{this.props.main}
					</div>
					{!(this.context as any).signedIn && <Footer />}
				</div>
				<EventListener target={global.document} event="sourcegraph:desktop" callback={this._handleSourcegraphDesktop} />
			</div>
		);
	}
}

export const rootRoute: ReactRouter.PlainRoute = {
	path: "/",
	component: withEventLoggerContext(EventLogger,
		withViewEventsLogged(
			withAppdashRouteStateRecording(
				withChannelListener(
					withSiteConfigContext(
						withUserContext(
							withFeaturesContext(
								redirectForDashboard(
									App
								)
							)
						)
					)
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
			...dashboardRoutes,
			...desktopRoutes,
			...homeRoutes,
			...channelRoutes,
			miscRoute,
			...adminRoutes,
			...searchRoutes,
			...userRoutes,
			...userSettingsRoutes,
			...repoRoutes,
		]);
	},
};
