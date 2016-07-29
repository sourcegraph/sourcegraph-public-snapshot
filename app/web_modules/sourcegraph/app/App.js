// @flow weak

import * as React from "react";
import Helmet from "react-helmet";
import type {Route} from "react-router";

import GlobalNav from "sourcegraph/app/GlobalNav";
import Footer from "sourcegraph/app/Footer";
import CSSModules from "react-css-modules";
import normalize from "sourcegraph/components/styles/_normalize.css"; // eslint-disable-line no-unused-vars
import styles from "./styles/App.css";

import {withEventLoggerContext, withViewEventsLogged} from "sourcegraph/util/EventLogger";
import EventLogger from "sourcegraph/util/EventLogger";
import {withFeaturesContext} from "sourcegraph/app/features";
import {withSiteConfigContext} from "sourcegraph/app/siteConfig";
import {withUserContext} from "sourcegraph/app/user";
import {withAppdashRouteStateRecording} from "sourcegraph/app/appdash";
import withChannelListener from "sourcegraph/channel/withChannelListener";
import desktopRouter from "sourcegraph/desktop";

const reactElement = React.PropTypes.oneOfType([
	React.PropTypes.arrayOf(React.PropTypes.element),
	React.PropTypes.element,
]);

export default class App extends React.Component {
	static propTypes = {
		main: reactElement,
		navContext: reactElement,
		globalNav: reactElement,
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
		let styleName = "main-container";
		if (!context.signedIn && location.pathname === "/") styleName = "main-container-homepage";
		this._handleSourcegraphDesktop = this._handleSourcegraphDesktop.bind(this);
		this.state = {
			styleName: styleName,
		};
	}

	state = {
		styleName: String,
	};

	componentDidMount() {
		if (typeof document !== "undefined") {
			document.addEventListener("sourcegraph:desktop", this._handleSourcegraphDesktop);
		}
	}

	componentWillUnmount() {
		document.removeEventListener("sourcegraph:desktop", this._handleSourcegraphDesktop);
	}

	_handleSourcegraphDesktop: any;
	_handleSourcegraphDesktop(event) {
		this.context.router.replace(event.detail);
	}

	render() {
		return (
			<div styleName={this.state.styleName}>
				<Helmet titleTemplate="%s · Sourcegraph" defaultTitle="Sourcegraph" />
				{this.props.globalNav || this.props.globalNav === null ?
					this.props.globalNav :
					<GlobalNav params={this.props.params} location={this.props.location} channelStatusCode={this.props.channelStatusCode}/>
				}
				<div styleName="main-content">
					{this.props.navContext && <div styleName="breadcrumb">{this.props.navContext}</div>}
					{this.props.main}
				</div>
				{!this.context.signedIn && <Footer />}
			</div>
		);
	}
}

export const rootRoute: Route = {
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
		require.ensure([], (require) => {
			callback(null, require("sourcegraph/dashboard").routes);
		});
	},
	getChildRoutes: (location, callback) => {
		require.ensure([], (require) => {
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
		});
	},
};
