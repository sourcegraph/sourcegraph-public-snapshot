// @flow

import React from "react";
import context from "sourcegraph/app/context";

// redirectIfLoggedIn wraps a component and issues a redirect
// if there is an authenticated user. It is useful for wrapping
// login, signup, etc., route components.
//
// TODO(sqs): would improve perf by just checking whether there's a
// valid accessToken (which indicates if the user is logged in); no
// need to wait for the whole user object to come in.
export default function redirectIfLoggedIn(url: Location | string, Component: ReactClass): ReactClass {
	class RedirectIfLoggedIn extends React.Component {
		static contextTypes = {
			router: React.PropTypes.object,
		};

		componentWillMount() {
			if (context.currentUser) this._redirect();
		}

		_redirect() {
			this.context.router.replace(url);
		}

		render() { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
