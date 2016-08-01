import * as React from "react";

// redirectIfLoggedIn wraps a component and issues a redirect
// if there is an authenticated user. It is useful for wrapping
// login, signup, etc., route components.
export default function redirectIfLoggedIn(url: Location | string, Component) {
	class RedirectIfLoggedIn extends React.Component {
		static contextTypes = {
			signedIn: React.PropTypes.bool.isRequired,
			router: React.PropTypes.object,
		};

		componentWillMount() {
			if (this.context.signedIn) this._redirect();
		}

		componentWillReceiveProps(nextProps, nextContext?: {signedIn: bool}) {
			if (nextContext && nextContext.signedIn) this._redirect();
		}

		_redirect() {
			this.context.router.replace(url);
		}

		render() { return <Component {...this.props} />; }
	}
	return RedirectIfLoggedIn;
}
