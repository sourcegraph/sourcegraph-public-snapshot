import React from "react";
import URL from "url";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import * as TreeActions from "sourcegraph/tree/TreeActions";

// All data from window.location gets processed here and is then passed down
// to sub-components via props. Every time window.location changes, this
// component gets re-rendered. Sub-components should never access
// window.location by themselves.
class TreeRouter extends Component {
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
			{
				const locationParts = this.state.location.split("/");
				locationParts.splice(locationParts.length - 1, 1);
				this.state.navigate(locationParts.join("/"));
				break;
			}

		case TreeActions.DownDirectory:
			{
				let url = URL.parse(this.state.location);
				if (url.pathname.indexOf(".tree") === -1) {
					url.pathname = `${this.state.repo}@${this.state.rev}/.tree/${action.part}`;
					// We are at the root of the directory tree; prefix /.tree on path.
					this.state.navigate(URL.format(url));
				} else {
					// Just append the part.
					let pathname = url.pathname;
					if (pathname[pathname.length - 1] !== "/") pathname = `${pathname}/`;
					url.pathname = `${pathname}${action.part}`;
					this.state.navigate(URL.format(url));
				}
				break;
			}

		case TreeActions.GoToDirectory:
			{
				let url = URL.parse(this.state.location);
				url.pathname = `${this.state.repo}@${this.state.rev}/.tree/${action.path.join("/")}`;
				this.state.navigate(URL.format(url));
				break;
			}
		}

	}

	_currPath() {
		const index = this.state.location.indexOf(".tree");
		if (index === -1) return [];

		const pathString = this.state.location.substring(index + ".tree/".length);
		if (pathString === "") return [];
		return pathString.split("/");
	}

	render() {
		return (
			<TreeSearch
				repo={this.state.repo}
				rev={this.state.rev}
				commitID={this.state.commitID}
				overlay={false}
				currPath={this._currPath()} />
		);
	}
}

TreeRouter.propTypes = {
	location: React.PropTypes.string.isRequired,
	navigate: React.PropTypes.func.isRequired,
	repo: React.PropTypes.string, // currently passed, but can (and should?) be inferred by the URL
	rev: React.PropTypes.string,  // same as above
	commitID: React.PropTypes.string,
};

export default TreeRouter;
