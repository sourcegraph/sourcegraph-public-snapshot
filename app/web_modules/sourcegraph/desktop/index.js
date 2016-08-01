import * as React from "react";
import {NotInBeta} from "sourcegraph/desktop/DesktopHome";
import desktopContainer from "sourcegraph/desktop/DesktopContainer";

import {rel} from "sourcegraph/app/routePatterns";
import {inBeta} from "sourcegraph/user";
import * as betautil from "sourcegraph/util/betautil";
import {getRouteName} from "sourcegraph/app/routePatterns";

export const desktopHome = {
	getComponent: (location, callback) => {
		require.ensure([], (require) => {
			callback(null, {
				main: require("sourcegraph/desktop/DesktopHome").default,
			});
		});
	},
};

export const routes = [
	{
		...desktopHome,
		path: rel.desktopHome,
	},
];

export default function desktopRouter(Component) {
	class DesktopRouter extends React.Component {
		static contextTypes = {
			router: React.PropTypes.object.isRequired,
			user: React.PropTypes.object,
			signedIn: React.PropTypes.bool.isRequired,
		};

		static propTypes = {
			routes: React.PropTypes.array,
		};

		render() {
			const inbeta = inBeta(this.context.user, betautil.DESKTOP);
			// Include this.context.user to prevent flicker when user loads
			if (this.context.signedIn && this.context.user && !inbeta) {
				return <NotInBeta />;
			}

			if (getRouteName(this.props.routes) === "home") {
				if (!this.context.signedIn) {
					// Prevent unauthed users from escaping
					this.context.router.replace(rel.login);
				} else {
					this.context.router.replace(rel.desktopHome);
				}
			}

			return <Component {...this.props} />;
		}
	}

	const DesktopClient = navigator.userAgent.includes("Electron");
	if (DesktopClient) {
		return desktopContainer(DesktopRouter);
	}
	return Component;
}
