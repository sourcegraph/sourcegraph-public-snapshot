// @flow

import React from "react";
import type {Route, RouteParams} from "react-router";
import {getViewName, getRoutePattern} from "./routePatterns";

import Component from "sourcegraph/Component";

import GlobalNav from "sourcegraph/app/GlobalNav";
import Footer from "sourcegraph/app/Footer";

import EventLogger from "sourcegraph/util/EventLogger";

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

export class App extends Component {
	static propTypes = {
		routes: React.PropTypes.arrayOf(React.PropTypes.object),
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
		this._logView(this.state.routes);
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
			this._logView(nextState.routes);
		}
	}

	_logView(routes: Array<Route>) {
		const viewName = getViewName(routes);
		if (viewName) {
			EventLogger.logEvent(viewName);
		} else {
			EventLogger.logEvent("UnmatchedRoute", {
				pattern: getRoutePattern(this.state.routes),
			});
		}
	}

	props: Props;

	render() {
		return (
			<div>
				<GlobalNav navContext={this.props.navContext} />
				{this.props.main}
				<Footer />
			</div>
		);
	}
}

export const rootRoute: Route = {
	path: "/",
	component: App,
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
