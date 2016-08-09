// tslint:disable

import * as React from "react";

// redirectForUnauthedUser wraps a component and issues a redirect
// if there is an unauthenticated user. It is useful for wrapping authed routes.
export default function redirectForUnauthedUser(url: Location | string, Component) {
	class RedirectForUnauthedUser extends React.Component<any, any> {
		static contextTypes = {
			signedIn: React.PropTypes.bool.isRequired,
			router: React.PropTypes.object.isRequired,
		};

		componentWillMount() {
			if (!(this.context as any).signedIn) this._redirect();
		}

		componentWillReceiveProps(nextProps, nextContext?: {signedIn: boolean}) {
			if (nextContext && !nextContext.signedIn) this._redirect();
		}

		_redirect() {
			(this.context as any).router.replace(url);
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectForUnauthedUser;
}
