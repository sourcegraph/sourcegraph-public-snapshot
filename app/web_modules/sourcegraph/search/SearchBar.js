import React from "react";
import ReactDOM from "react-dom";
import URI from "urijs";

import Component from "sourcegraph/Component";
import LocationAdaptor from "sourcegraph/LocationAdaptor";
import SearchResultsRouter from "sourcegraph/search/SearchResultsRouter";

class SearchBar extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchViewIsActive: false,
		};
		this._submitSearch = this._submitSearch.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.uri = URI.parse(props.location);
		state.navigate = props.navigate || null;

		let pathParts = state.uri.path.substr(1).split("/.");
		state.searchViewIsActive = (pathParts.length >= 2 && pathParts[1] === "search");

		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0] || null;
		state.rev = repoParts[1] || "master";

		// The "~" prefix indicates we're on a user page, not a repo.
		if (state.repo && state.repo[0] === "~") {
			state.repo = null;
			state.rev = null;
		}

		let vars = URI.parseQuery(state.uri.query);
		state.query = vars["q"] || null;
		state.type = vars["type"] || null;
	}

	onStateTransition(prevState, nextState) {
		if (!prevState.searchViewIsActive && nextState.searchViewIsActive) {
			let el;
			let repoName = nextState.repo.split("/").pop();
			// TODO(autotest) support document object.
			if (typeof document !== "undefined") document.title = `Search - ${repoName} - Sourcegraph`;
			if (typeof document !== "undefined") el = document.getElementById("main");
			if (el) {
				ReactDOM.render((
					<LocationAdaptor component={SearchResultsRouter} />
				), el);
			}
		} else if (prevState.searchViewIsActive && !nextState.searchViewIsActive) {
			window.location.href = URI.build(nextState.uri);
		}

		if (prevState.query !== nextState.query && this.refs.queryInput) {
			this.refs.queryInput.value = nextState.query;
		}
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

	_searchPath() {
		let revPart = this.state.rev ? `@${this.state.rev}` : "";
		return `/${this.state.repo}${revPart}/.search`;
	}

	_submitSearch(e) {
		if (e) e.preventDefault();
		// TODO(autotest) support refs.
		let query = this.refs.queryInput && this.refs.queryInput.value;
		if (!query) return;
		this._navigate(this._searchPath(), {
			q: query,
			type: this.state.type || "tokens",
			page: 1,
		});
	}

	render() {
		// Only search within a repo for now.
		if (!this.state.repo) return null;

		return (
			<form className="navbar-form" onSubmit={this._submitSearch.bind(this)}>
				<div className="form-group">
					<div className="input-group">
						<input className="form-control search-input"
							ref="queryInput"
							name="q"
							placeholder="Search"
							defaultValue={this.state.query} />
						<span className="input-group-addon"></span>
					</div>
				</div>
			</form>
		);
	}
}

export default SearchBar;
