// @flow

import React from "react";
import type {Route, RouteParams} from "react-router";
import GlobalNav from "sourcegraph/app/GlobalNav";
import Footer from "sourcegraph/app/Footer";

const reactElement = React.PropTypes.oneOfType([
	React.PropTypes.arrayOf(React.PropTypes.element),
	React.PropTypes.element,
]);

type Props = {
	children: Array<any>;
	main: Array<any>;
	navContext: Array<any>;
	params: RouteParams;
}

export class App extends React.Component {
	static propTypes = {
		children: reactElement,
		main: reactElement,
		navContext: reactElement,
		params: React.PropTypes.object,
	};
	static defaultProps: {};
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
