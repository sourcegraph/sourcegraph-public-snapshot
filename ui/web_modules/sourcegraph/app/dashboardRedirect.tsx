import * as React from "react";

import {abs} from "sourcegraph/app/routePatterns";
import {redirectForDesktop} from "sourcegraph/desktop/index";

export function redirectForDashboard<P>(Component: React.ComponentClass<P>): React.ComponentClass<P> {
	type Props = {
		location: {pathname: string},
	};

	class RedirectForDashboard extends React.Component<Props, any> {
		static contextTypes: React.ValidationMap<any> = {
			signedIn: React.PropTypes.bool.isRequired,
			router: React.PropTypes.object.isRequired,
		};

		context: {
			router: {replace: any},
			signedIn: boolean,
		};

		componentWillMount(): void {
			if (this.context.signedIn) {
				this._redirect();
			}
		}

		componentWillReceiveProps(nextProps: Props, nextContext?: {signedIn: boolean}): void {
			if (nextContext && nextContext.signedIn) {
				this._redirect();
			}
		}

		_redirect(): void {
			if (location.pathname === "/") {
				this.context.router.replace(`/${abs.dashboard}`);
			}
		}

		render(): JSX.Element | null { return <Component {...this.props} />; }
	}
	return redirectForDesktop(RedirectForDashboard);
}
