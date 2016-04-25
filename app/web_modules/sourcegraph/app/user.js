// @flow

import React from "react";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import UserStore from "sourcegraph/user/UserStore";
import "sourcegraph/user/UserBackend";
import * as UserActions from "sourcegraph/user/UserActions";
import type {User} from "sourcegraph/user";

// withUserContext passes user-related context items
// to Component's children.
export function withUserContext(Component: ReactClass): ReactClass {
	class WithUser extends Container {
		static childContextTypes = {
			user: React.PropTypes.object,

			// signedIn is knowable without hitting the network, so components
			// that only care "is there a logged-in user?" should use signedIn,
			// not `user !== null`, to check for that.
			signedIn: React.PropTypes.bool.isRequired,

			// githubToken is the user's ExternalToken for github.com.
			githubToken: React.PropTypes.object,
		};

		constructor(props) {
			super(props);
		}

		stores() { return [UserStore]; }

		reconcileState(state, props) {
			Object.assign(state, props);

			state.accessToken = UserStore.activeAccessToken || null;
			state.authInfo = state.accessToken ? UserStore.authInfo.get(state.accessToken) : null;
			state.githubToken = UserStore.activeGitHubToken || null;
			state.user = state.authInfo && !state.authInfo.Error ? UserStore.users.get(state.authInfo.UID) : null;
		}

		onStateTransition(prevState, nextState) {
			if (nextState.accessToken && prevState.accessToken !== nextState.accessToken) {
				Dispatcher.Backends.dispatch(new UserActions.WantAuthInfo(nextState.accessToken));
			}

			if (prevState.authInfo !== nextState.authInfo) {
				if (nextState.authInfo && !nextState.authInfo.Error && nextState.authInfo.UID) {
					Dispatcher.Backends.dispatch(new UserActions.WantUser(nextState.authInfo.UID));
				}
			}
		}

		getChildContext(): {user: ?User} {
			return {
				user: this.state.user && !this.state.user.Error ? this.state.user : null,
				signedIn: Boolean(this.state.accessToken),
				githubToken: this.state.githubToken,
			};
		}

		render() {
			return <Component {...this.props} />;
		}
	}
	return WithUser;
}
