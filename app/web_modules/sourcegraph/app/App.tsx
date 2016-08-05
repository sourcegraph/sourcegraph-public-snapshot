// tslint:disable

import * as React from "react";
import Helmet from "react-helmet";
import {Route} from "react-router";

import GlobalNav from "sourcegraph/app/GlobalNav";
import Footer from "sourcegraph/app/Footer";
import CSSModules from "react-css-modules";
import "sourcegraph/components/styles/_normalize.css";
import styles from "./styles/App.css";

import {withEventLoggerContext, withViewEventsLogged} from "sourcegraph/util/EventLogger";
import EventLogger from "sourcegraph/util/EventLogger";
import {withFeaturesContext} from "sourcegraph/app/features";
import {withSiteConfigContext} from "sourcegraph/app/siteConfig";
import {withUserContext} from "sourcegraph/app/user";
import {withAppdashRouteStateRecording} from "sourcegraph/app/appdash";
import withChannelListener from "sourcegraph/channel/withChannelListener";
import desktopRouter from "sourcegraph/desktop/index";

const reactElement = React.PropTypes.oneOfType([
	React.PropTypes.arrayOf(React.PropTypes.element),
	React.PropTypes.element,
]);

export default class App extends React.Component<any, any> {
	static propTypes = {
		main: reactElement,
		navContext: reactElement,
		location: React.PropTypes.object.isRequired,
		params: React.PropTypes.object,
		channelStatusCode: React.PropTypes.number,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
	};

	constructor(props, context) {
		super(props);
		let styleName = "main_container";
		if (!context.signedIn && location.pathname === "/") styleName = "main_container_homepage";
		this._handleSourcegraphDesktop = this._handleSourcegraphDesktop.bind(this);
		this.state = {
			styleName: styleName,
		};
	}

	state = {
		styleName: "",
	};

	componentDidMount() {
		if (typeof document !== "undefined") {
			document.addEventListener("sourcegraph:desktop", this._handleSourcegraphDesktop);
		}
	}

	componentWillUnmount() {
		document.removeEventListener("sourcegraph:desktop", this._handleSourcegraphDesktop);
	}

	_handleSourcegraphDesktop(event) {
		(this.context as any).router.replace(event.detail);
	}

	render(): JSX.Element | null {
		return (
			<div styleName={this.state.styleName}>
				<Helmet titleTemplate="%s · Sourcegraph" defaultTitle="Sourcegraph" />
				<GlobalNav params={this.props.params} location={this.props.location} channelStatusCode={this.props.channelStatusCode}/>
				<div styleName="main_content">
					{this.props.navContext && <div styleName="breadcrumb">{this.props.navContext}</div>}
					{this.props.main}
				</div>
				{!(this.context as any).signedIn && <Footer />}
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
                            desktopRouter(
                                withFeaturesContext(
                                    CSSModules(App, styles)
                                )
                            )
                        )
                    )
				)
			)
		)
	),
	getIndexRoute: (location, callback) => {
		callback(null, require("sourcegraph/dashboard").routes);
	},
	getChildRoutes: (location, callback) => {
		callback(null, [
			...require("sourcegraph/page").routes,
			...require("sourcegraph/styleguide").routes,
			...require("sourcegraph/desktop").routes,
			...require("sourcegraph/home").routes,
			...require("sourcegraph/channel").routes,
			require("sourcegraph/misc/golang").route,
			...require("sourcegraph/admin/routes").routes,
			...require("sourcegraph/search/routes").routes,
			...require("sourcegraph/user").routes,
			...require("sourcegraph/user/settings/routes").routes,
			...require("sourcegraph/repo/routes").routes,
		]);
	},
};
