// tslint:disable typedef ordered-imports
import * as React from "react";
import {rel} from "sourcegraph/app/routePatterns";
import {DashboardContainer} from "sourcegraph/dashboard/DashboardContainer";
import {Home} from "sourcegraph/home/Home";
import {DesktopHome} from "sourcegraph/desktop/DesktopHome";
import {IntegrationsContainer} from "sourcegraph/home/IntegrationsContainer";
import {context} from "sourcegraph/app/context";

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
	render() {
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
