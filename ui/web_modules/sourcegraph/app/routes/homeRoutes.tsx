import * as React from "react";
import {context} from "sourcegraph/app/context";
import {rel} from "sourcegraph/app/routePatterns";
import {Dashboard} from "sourcegraph/dashboard/Dashboard";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Home} from "sourcegraph/home/Home";
import {IntegrationsContainer} from "sourcegraph/home/IntegrationsContainer";
import * as OrgActions from "sourcegraph/org/OrgActions";

export const homeRoutes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: HomeRouter,
			});
		},
		path: rel.home,
	},
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: IntegrationsContainer,
			});
		},
		path: rel.integrations,
	},
];

class HomeRouter extends React.Component<any, null> {
	componentDidMount(): void {
		// Fetch authed user organizations on the dashboard to update user properties.
		if (context.user && context.hasOrganizationGitHubToken()) {
			Dispatcher.Backends.dispatch(new OrgActions.WantOrgs(context.user.Login));
		}
	}

	render(): JSX.Element | null {
		if (context.user) {
			return <Dashboard {...this.props}/>;
		}
		return <Home {...this.props} />;
	}
}
