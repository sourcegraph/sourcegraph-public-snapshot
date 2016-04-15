// @flow

import React from "react";
import Helmet from "react-helmet";
import type {Route, RouteParams} from "react-router";
import {getViewName, getRoutePattern} from "./routePatterns";

import Component from "sourcegraph/Component";

import GlobalNav from "sourcegraph/app/GlobalNav";
import Footer from "sourcegraph/app/Footer";
import CSSModules from "react-css-modules";
import styles from "./styles/App.css";

import EventLogger from "sourcegraph/util/EventLogger";
import {withStatusContext} from "sourcegraph/app/status";
import context from "sourcegraph/app/context";

const reactElement = React.PropTypes.oneOfType([
	React.PropTypes.arrayOf(React.PropTypes.element),
	React.PropTypes.element,
]);

type Props = {
	routes: Array<Route>;
	location: Object;
	children: Array<any>;
	main: Array<any>;
	navContext: Array<any>;
	params: RouteParams;
};

type State = {
	routes: Array<Route>;
	location: Object;
	children: Array<any>;
	main: Array<any>;
	navContext: Array<any>;
	params: RouteParams;
	routePattern: string;
};

class App extends Component {
	static propTypes = {
		routes: React.PropTypes.arrayOf(React.PropTypes.object),
		route: React.PropTypes.object.isRequired,
		location: React.PropTypes.object,
		children: reactElement,
		main: reactElement,
		navContext: reactElement,
		params: React.PropTypes.object,
	};
	static defaultProps: {};

	constructor(props: Props) {
		super(props);
		this._hasMounted = false;
		EventLogger.init();
	}

	componentDidMount() {
		this._hasMounted = true;
		this._logView(this.state.routes, this.state.location);
	}

	reconcileState(state: State, props: Props) {
		Object.assign(state, props);
		state.routePattern = props.routes.map((route) => route.path).join("").slice(1); // remove leading '/'
	}

	onStateTransition(prevState: State, nextState: State) {
		if (this._hasMounted && prevState.location.pathname !== nextState.location.pathname) {
			// Greedily log page views. Technically changing the pathname
			// may match the same "view" (e.g. interacting with the directory
			// tree navigations will change your URL,  but not feel like separate
			// page events). We will log any change in pathname as a separate event.
			// NOTE: this will not log separate page views when query string / hash
			// values are updated.
			this._logView(nextState.routes, nextState.location);
		}
	}

	_logView(routes: Array<Route>, location: Object) {
		let eventProps = {
			referred_by_chrome_ext: false,
			url: location.pathname,
		};
		if (location.query && location.query["utm_source"] === "chromeext") {
			eventProps.referred_by_chrome_ext = true;
		}

		const viewName = getViewName(routes);
		if (viewName) {
			EventLogger.logEvent(viewName, eventProps);
		} else {
			EventLogger.logEvent("UnmatchedRoute", {
				...eventProps,
				pattern: getRoutePattern(routes),
			});
		}
	}

	props: Props;

	render() {
		return (
			<div styleName="main-container">
				<Helmet titleTemplate="%s · Sourcegraph" defaultTitle="Sourcegraph" />
				{(context.currentUser || this.state.location.pathname !== "/") && <GlobalNav navContext={this.props.navContext} />}
				<div styleName="main-content">{this.props.main}</div>
				<Footer />
			</div>
		);
	}
}

export const rootRoute: Route = {
	path: "/",
	component: withStatusContext(CSSModules(App, styles)),
	getIndexRoute: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, require("sourcegraph/dashboard").route);
		});
	},
	getChildRoutes: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, [
				...require("sourcegraph/admin/routes").routes,
				...require("sourcegraph/user").routes,
				...require("sourcegraph/repo/routes").routes,
			]);
		});
	},
};
