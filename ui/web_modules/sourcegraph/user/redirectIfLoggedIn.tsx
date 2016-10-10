// tslint:disable: typedef ordered-imports

import * as React from "react";
import {InjectedRouter} from "react-router";
import {context} from "sourcegraph/app/context";

// redirectIfLoggedIn wraps a component and issues a redirect
// if there is an authenticated user. It is useful for wrapping
// login, signup, etc., route components.
//
// TODO: remove queryObj overriding for onboarding step.
export function redirectIfLoggedIn(url: Location | string, queryObj: History.Query, Component) {
	type Props = any;

	type State = any;

	class RedirectIfLoggedIn extends React.Component<Props, State> {
		static contextTypes: React.ValidationMap<any> = {
			router: React.PropTypes.object.isRequired,
		};

		context: {
			router: InjectedRouter;
		};

		componentWillMount(): void {
			const redirQueryObj = Object.assign({}, this.props.location.query || null, queryObj);
			const redirRouteObj = typeof url === "string" ? {pathname: url} : url;
			const redirLocation = Object.assign({}, this.props.location || null, redirRouteObj, {query: redirQueryObj});

			if (context.user) {
				this.context.router.replace(redirLocation);
			}
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
