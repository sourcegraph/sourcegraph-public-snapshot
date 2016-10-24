import * as React from "react";
import {context} from "sourcegraph/app/context";
import {rel} from "sourcegraph/app/routePatterns";
import {DashboardContainer} from "sourcegraph/dashboard/DashboardContainer";
import {DesktopHome} from "sourcegraph/desktop/DesktopHome";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Home} from "sourcegraph/home/Home";
import {IntegrationsContainer} from "sourcegraph/home/IntegrationsContainer";
import * as OrgActions from "sourcegraph/org/OrgActions";

export const routes: any[] = [
	{
		getComponent: (location, callback) => {
			callback(null, {
				main: HomeRouter,
				navContext: null,
			});
		},
		path: rel.home,
	},
	{
		getComponent: (location, callback) => {
			callback(null, {
				navContext: null,
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
		const desktopClient = navigator.userAgent.includes("Electron");
		if (desktopClient) {
			return <DesktopHome />;
		}
		if (context.user) {
			return <DashboardContainer {...this.props}/>;
		}
		return <Home {...this.props} />;
	}
}
