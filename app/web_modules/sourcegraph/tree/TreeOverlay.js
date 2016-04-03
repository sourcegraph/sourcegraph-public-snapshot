import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import * as TreeActions from "sourcegraph/tree/TreeActions";

class TreeOverlay extends Component {
	constructor(props) {
		super(props);
		this.state = {
			path: [],
		};
	}

	componentDidMount() {
		this.dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		Dispatcher.unregister(this.dispatcherToken);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case TreeActions.UpDirectory:
			this.setState({
				path: update(this.state.path, {$splice: [[this.state.path.length - 1, 1]]}),
			});
			break;

		case TreeActions.DownDirectory:
			this.setState({
				path: update(this.state.path, {$push: [action.part]}),
			});
			break;

		case TreeActions.GoToDirectory:
			this.setState({
				path: action.path,
			});
			break;
		}
	}

	render() {
		return (
			<TreeSearch
				repo={this.state.repo}
				rev={this.state.rev}
				overlay={true}
				prefetch={false}
				path={this.state.path} />
		);
	}
}

TreeOverlay.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
};

export default TreeOverlay;
