// tslint:disable typedef ordered-imports
import * as React from "react";
import {rel} from "sourcegraph/app/routePatterns";
import {DashboardContainer} from "sourcegraph/dashboard/DashboardContainer";
import {Home} from "sourcegraph/home/Home";
import {DesktopHome} from "sourcegraph/desktop/DesktopHome";
import {IntegrationsContainer} from "sourcegraph/home/IntegrationsContainer";

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
	static contextTypes: React.ValidationMap<any> = {
		signedIn: React.PropTypes.bool.isRequired,
	};

	context: {signedIn: boolean};
	render() {
		const desktopClient = navigator.userAgent.includes("Electron");
		if (desktopClient) {
			return <DesktopHome />;
		}
		if (this.context.signedIn) {
			return <DashboardContainer {...this.props}/>;
		}
		return <Home {...this.props} />;
	}
}
