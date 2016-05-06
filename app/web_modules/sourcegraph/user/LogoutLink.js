import React from "react";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import UserStore from "sourcegraph/user/UserStore";
import "sourcegraph/user/UserBackend"; // for side effects

export default class LogoutLink extends Container {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	state = {
		submitted: false,
	};

	constructor(props) {
		super(props);
		this._handleClick = this._handleClick.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.pendingAuthAction = UserStore.pendingAuthActions.get("logout");
		state.authResponse = UserStore.authResponses.get("logout");
	}

	onStateTransition(prevState, nextState) {
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
		});
	}

	render() {
		return (
			<a {...this.props} onClick={this._handleClick}>
				{this.state.submitted && (this.state.pendingAuthAction || (this.state.authResponse && !this.state.authResponse.Error)) ? "..." : "Sign out"}
			</a>
		);
	}
}

