// tslint:disable: typedef ordered-imports curly

import * as React from "react";

// redirectIfLoggedIn wraps a component and issues a redirect
// if there is an authenticated user. It is useful for wrapping
// login, signup, etc., route components.
export function redirectIfLoggedIn(url: Location | string, Component) {
	class RedirectIfLoggedIn extends React.Component<any, any> {
		static contextTypes = {
			signedIn: React.PropTypes.bool.isRequired,
			router: React.PropTypes.object.isRequired,
		};

		componentWillMount() {
			if ((this.context as any).signedIn) this._redirect();
		}

		componentWillReceiveProps(nextProps, nextContext?: {signedIn: boolean}) {
			if (nextContext && nextContext.signedIn) this._redirect();
		}

		_redirect() {
			(this.context as any).router.replace(url);
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
