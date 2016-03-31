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
			currPath: [],
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
				currPath: update(this.state.currPath, {$splice: [[this.state.currPath.length - 1, 1]]}),
			});
			break;

		case TreeActions.DownDirectory:
			this.setState({
				currPath: update(this.state.currPath, {$push: [action.part]}),
			});
			break;

		case TreeActions.GoToDirectory:
			this.setState({
				currPath: action.path,
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
				prefetch={true}
				currPath={this.state.currPath} />
		);
	}
}

TreeOverlay.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
};

export default TreeOverlay;
