import React from "react";

// redirectForUnauthedUser wraps a component and issues a redirect
// if there is an unauthenticated user. It is useful for wrapping authed routes.
export default function redirectForUnauthedUser(url: Location | string, Component: ReactClass): ReactClass {
	class RedirectForUnauthedUser extends React.Component {
		static contextTypes = {
			signedIn: React.PropTypes.bool.isRequired,
			router: React.PropTypes.object,
		};

		componentWillMount() {
			if (!this.context.signedIn) this._redirect();
		}

		componentWillReceiveProps(nextProps, nextContext?: {signedIn: bool}) {
			if (nextContext && !nextContext.signedIn) this._redirect();
		}

		_redirect() {
			this.context.router.replace(url);
		}

		render() { return <Component {...this.props} />; }
	}
	return RedirectForUnauthedUser;
}
