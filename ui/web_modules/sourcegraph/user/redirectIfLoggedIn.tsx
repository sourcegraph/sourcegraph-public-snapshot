// tslint:disable: typedef ordered-imports

import * as React from "react";
import {InjectedRouter} from "react-router";
import {context} from "sourcegraph/app/context";

// redirectIfLoggedIn wraps a component and issues a redirect
// if there is an authenticated user. It is useful for wrapping
// login, signup, etc., route components.
export function redirectIfLoggedIn(url: Location | string, Component) {
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
			if (context.user) {
				this.context.router.replace(url);
			}
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
