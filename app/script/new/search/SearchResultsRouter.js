import React from "react";
import URI from "urijs";

import Component from "../Component";
import Dispatcher from "../Dispatcher";
import * as SearchActions from "./SearchActions";
import SearchResultsContainer from "./SearchResultsContainer";

class SearchResultsRouter extends Component {
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
		state.type = vars["type"] || "tokens";
		state.page = parseInt(vars["page"], 10) || 1;
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
				page: 1,
			});
			break;

		case SearchActions.SelectPage:
			this._navigate(this._searchPath(), {
				page: action.page,
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
				query={this.state.query}
				type={this.state.type}
				page={this.state.page} />
		);
	}
}

SearchResultsRouter.propTypes = {
	location: React.PropTypes.string,
	navigate: React.PropTypes.func,
};

export default SearchResultsRouter;
