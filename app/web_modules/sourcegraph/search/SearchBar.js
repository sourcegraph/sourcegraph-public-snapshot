import React from "react";
import ReactDOM from "react-dom";
import URL from "url";

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
		state.url = URL.parse(props.location, true);
		state.navigate = props.navigate || null;

		let pathParts = state.url.pathname.substr(1).split("/.");
		state.searchViewIsActive = (pathParts.length >= 2 && pathParts[1] === "search");

		let repoParts = pathParts[0].split("@");
		state.repo = repoParts[0] || null;
		state.rev = repoParts[1] || "master";

		// The "~" prefix indicates we're on a user page, not a repo.
		if (state.repo && state.repo[0] === "~") {
			state.repo = null;
			state.rev = null;
		}

		state.query = state.url.query["q"] || null;
		state.type = state.url.query["type"] || null;
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
			window.location.href = URL.format(nextState.url);
		}

		if (prevState.query !== nextState.query && this.refs.queryInput) {
			this.refs.queryInput.value = nextState.query;
		}
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
