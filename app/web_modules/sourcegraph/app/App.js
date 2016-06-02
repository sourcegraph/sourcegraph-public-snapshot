// @flow

import React from "react";
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

const reactElement = React.PropTypes.oneOfType([
	React.PropTypes.arrayOf(React.PropTypes.element),
	React.PropTypes.element,
]);

function App(props, {signedIn}) {
	let styleName = "main-container";
	if (props.location.state && props.location.state.modal) styleName = "main-container-with-modal";
	if (!signedIn && location.pathname === "/") styleName = "main-container-homepage";

	return (
		<div styleName={styleName}>
			<Helmet titleTemplate="%s · Sourcegraph" defaultTitle="Sourcegraph" />
			<GlobalNav location={props.location} channelStatusCode={props.channelStatusCode}/>
			<div styleName="main-content">
				{props.navContext && <div styleName="breadcrumb">{props.navContext}</div>}
				{props.main}
			</div>
			{!signedIn && <Footer />}
		</div>
	);
}
App.propTypes = {
	main: reactElement,
	navContext: reactElement,
	location: React.PropTypes.object.isRequired,
	channelStatusCode: React.PropTypes.number,
};

App.contextTypes = {
	signedIn: React.PropTypes.bool.isRequired,
};

export const rootRoute: Route = {
	path: "/",
	component: withEventLoggerContext(EventLogger,
		withViewEventsLogged(
			withAppdashRouteStateRecording(
				withChannelListener(
					withSiteConfigContext(
						withUserContext(
							withFeaturesContext(
								CSSModules(App, styles)
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
				...require("sourcegraph/home").routes,
				...require("sourcegraph/channel").routes,
				require("sourcegraph/misc/golang").route,
				...require("sourcegraph/admin/routes").routes,
				...require("sourcegraph/search/routes").routes,
				...require("sourcegraph/user").routes,
				...require("sourcegraph/repo/routes").routes,
			]);
		});
	},
};
