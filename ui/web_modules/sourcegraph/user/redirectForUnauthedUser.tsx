// tslint:disable: typedef ordered-imports

import * as React from "react";

type Props = any;

type State = any;

// redirectForUnauthedUser wraps a component and issues a redirect
// if there is an unauthenticated user. It is useful for wrapping authed routes.
export function redirectForUnauthedUser(url: Location | string, Component) {
	class RedirectForUnauthedUser extends React.Component<Props, State> {
		static contextTypes: React.ValidationMap<any> = {
			signedIn: React.PropTypes.bool.isRequired,
			router: React.PropTypes.object.isRequired,
		};

		componentWillMount(): void {
			if (!(this.context as any).signedIn) {
				this._redirect();
			}
		}

		componentWillReceiveProps(nextProps: Props, nextContext?: {signedIn: boolean}) {
			if (nextContext && !nextContext.signedIn) {
				this._redirect();
			}
		}

		_redirect() {
			(this.context as any).router.replace(url);
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return RedirectForUnauthedUser;
}
