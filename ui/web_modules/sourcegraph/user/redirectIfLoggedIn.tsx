import { History } from "history";
import * as React from "react";

import { context } from "sourcegraph/app/context";
import { Router, RouterLocation } from "sourcegraph/app/router";

interface Props {
	location: RouterLocation;
}

// redirectIfLoggedIn wraps a component and issues a redirect
// if there is an authenticated user. It is useful for wrapping
// login, signup, etc., route components.
//
// TODO: remove queryObj overriding for onboarding step.
export function redirectIfLoggedIn(url: Location | string, queryObj: History.Query, Component: React.ReactType): React.ComponentClass<Props> {

	class RedirectIfLoggedIn extends React.Component<Props, {}> {
		static contextTypes: React.ValidationMap<any> = {
			router: React.PropTypes.object.isRequired,
		};

		context: {
			router: Router;
		};

		componentWillMount(): void {
			const redirQueryObj = Object.assign({}, queryObj, this.props.location.query || null);
			const redirRouteObj = typeof url === "string" ? { pathname: url } : url;
			const redirLocation = Object.assign({}, this.props.location || null, redirRouteObj, { query: redirQueryObj });

			if (context.user) {
				this.context.router.replace(redirLocation);
			}
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
