import React from "react";
import URI from "urijs";

import Component from "../Component";
import Dispatcher from "../Dispatcher";
import * as SearchActions from "./SearchActions";
import SearchResultsContainer from "./SearchResultsContainer";

export default class SearchResultsRouter extends Component {
	componentDidMount() {
		this.dispatcherToken = Dispatcher.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		Dispatcher.unregister(this.dispatcherToken);
	}

	reconcileState(state, props) {
		state.uri = URI.parse(props.location);
		state.navigate = props.navigate || null;

		let pathParts = state.uri.path.substr(1).split("/.");
		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0];
		state.rev = repoParts[1] || "master";

		let vars = URI.parseQuery(state.uri.query);
		state.query = vars["q"] || null;
		state.type = vars["type"] || null;
	}

	_navigate(path, query) {
		let uri = Object.assign({}, this.state.uri);
		if (path) {
			uri.path = path;
		}
		if (query) {
			uri.query = URI.buildQuery(Object.assign(URI.parseQuery(uri.query), query));
		}
		this.state.navigate(URI.build(uri));
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case SearchActions.SelectResultType:
			this._navigate(this._searchPath(), {
				type: action.type,
			});
			break;
		}
	}

	_searchPath() {
		return `/${this.state.repo}@${this.state.rev}/.search`;
	}

	render() {
		return (
			<SearchResultsContainer
				repo={this.state.repo}
				rev={this.state.rev}
				type={this.state.type}
				query={this.state.query} />
		);
	}
}

SearchResultsRouter.propTypes = {
	location: React.PropTypes.string,
	navigate: React.PropTypes.func,
};
