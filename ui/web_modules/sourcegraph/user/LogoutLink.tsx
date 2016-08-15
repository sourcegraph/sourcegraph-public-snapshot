// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import {UserStore} from "sourcegraph/user/UserStore";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	role: string;
}

type State = any;

export class LogoutLink extends Container<Props, State> {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props: Props) {
		super(props);
		this._handleClick = this._handleClick.bind(this);
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions["logout"] || false;
		state.authResponse = UserStore.authResponses["logout"] || null;
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState.authResponse !== nextState.authResponse) {
			if (this.state.submitted) {
				if (nextState.authResponse && nextState.authResponse.Error) {
					console.error(`Logout failed: ${nextState.authResponse.Error.body.message}`);
				}
			}
		}
	}

	stores() { return [UserStore]; }

	_handleClick(ev) {
		ev.preventDefault();
		this.setState({submitted: true}, () => {
			Dispatcher.Stores.dispatch(new UserActions.SubmitLogout());
			Dispatcher.Backends.dispatch(new UserActions.SubmitLogout());

			const settings = UserStore.settings;
			const newSettings = Object.assign({}, settings, {
				search: Object.assign({}, settings.search, {
					scope: Object.assign({}, settings.search && settings.search.scope, {
						popular: true,
					}),
				}),
			});

			Dispatcher.Stores.dispatch(new UserActions.UpdateSettings(newSettings));
		});
	}

	render(): JSX.Element | null {
		return (
			<a {...this.props} onClick={this._handleClick}>
				{this.state.submitted && (this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)) ? "..." : "Sign out"}
			</a>
		);
	}
}
