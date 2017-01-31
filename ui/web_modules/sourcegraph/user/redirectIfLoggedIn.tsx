import * as React from "react";

import { context } from "sourcegraph/app/context";
import { Router, RouterLocation } from "sourcegraph/app/router";

interface Props {
	location: RouterLocation;
}

/**
 * redirectIfLoggedIn wraps a component and issues a redirect
 * if there is an authenticated user. It is useful for wrapping
 * login, signup, etc., route components.
 */
export function redirectIfLoggedIn(url: Location | string, Component: React.ReactType): React.ComponentClass<Props> {

	class RedirectIfLoggedIn extends React.Component<Props, {}> {
		static contextTypes: React.ValidationMap<any> = {
			router: React.PropTypes.object.isRequired,
		};

		context: {
			router: Router;
		};

		componentWillMount(): void {
			const redirRouteObj = typeof url === "string" ? { pathname: url } : url;
			const redirLocation = Object.assign({}, this.props.location || null, redirRouteObj);

			if (context.user) {
				if (!context.hasPrivateGitHubToken() && this.context.router.location.query["private"]) {
					// short-circuit the redirect, we want to allow the user to upgrade their token
					// TODO(john,dadlerj): this should be replaced with an upgrade page
				} else {
					this.context.router.replace(redirLocation);
				}
			}
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
