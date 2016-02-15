import React from "react";
import URL from "url";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import * as SearchActions from "sourcegraph/search/SearchActions";
import SearchResultsContainer from "sourcegraph/search/SearchResultsContainer";

class SearchResultsRouter extends Component {
	componentDidMount() {
		this.dispatcherToken = Dispatcher.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		Dispatcher.unregister(this.dispatcherToken);
	}

	reconcileState(state, props) {
		state.url = URL.parse(props.location, true);
		state.navigate = props.navigate || null;

		let pathParts = state.url.pathname.substr(1).split("/.");
		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0];
		state.rev = repoParts[1] || "master";

		state.query = state.url.query["q"] || null;
		state.type = state.url.query["type"] || "tokens";
		state.page = parseInt(state.url.query["page"], 10) || 1;
	}

	_navigate(pathname, query) {
		let url = {
			protocol: this.state.url.protocol,
			auth: this.state.url.auth,
			host: this.state.url.host,
			pathname: pathname || this.state.url.pathname,
			query: Object.assign({}, this.state.url.query, query),
		};
		this.state.navigate(URL.format(url));
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
		let revPart = this.state.rev ? `@${this.state.rev}` : "";
		return `/${this.state.repo}${revPart}/.search`;
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
