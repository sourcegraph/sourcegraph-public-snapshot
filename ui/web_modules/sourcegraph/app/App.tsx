import "sourcegraph/components/styles/_normalize.css";

import * as React from "react";
import { PlainRoute } from "react-router";

import { context } from "sourcegraph/app/context";
import { GlobalNav } from "sourcegraph/app/GlobalNav";
import { rel } from "sourcegraph/app/routePatterns";
import { Router, setRouter } from "sourcegraph/app/router";
import { pageRoutes } from "sourcegraph/app/routes/pageRoutes";
import { repoRoutes } from "sourcegraph/app/routes/repoRoutes";
import { styleguideRoutes } from "sourcegraph/app/routes/styleguideRoutes";
import { symbolRoutes } from "sourcegraph/app/routes/symbolRoutes";
import { userRoutes } from "sourcegraph/app/routes/userRoutes";
import { userSettingsRoutes } from "sourcegraph/app/routes/userSettingsRoutes";
import * as styles from "sourcegraph/app/styles/App.css";
import { HomeRouter } from "sourcegraph/home/HomeRouter";
import { withViewEventsLogged } from "sourcegraph/tracking/withViewEventsLogged";

interface Props {
	main: JSX.Element;
	footer?: JSX.Element;
}

export class App extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	constructor(props: Props, context: { router: Router }) {
		super(props, context);
		setRouter(context.router);
	}

	render(): JSX.Element {
		let className = styles.main_container;
		if (!context.user && this.context.router.location.pathname === "/") {
			className = styles.main_container_homepage;
		}
		return (
			<div className={className}>
				<GlobalNav />
				{this.props.main}
				{this.props.footer}
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
				cb(null, { main: HomeRouter });
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
